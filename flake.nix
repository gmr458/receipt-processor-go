{
  description = "Development environment for receipt-processor-go";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      # `nix fmt` formats every .nix file in the repo; treefmt-wrapped so it
      # works with no arguments.
      formatter = forAllSystems (pkgs: pkgs.nixfmt-tree);

      devShells = forAllSystems (
        pkgs:
        let
          # nixpkgs' go-migrate compiles in every DB driver; the bundled
          # snowflake driver panics in init() ("failed to parse CA
          # certificate") in a nix store build. Only the postgres/pgx tags
          # this project needs are kept. Build tags don't affect vendor
          # hashes, so this stays on nixpkgs' pinned sources.
          goMigrate = pkgs.go-migrate.overrideAttrs (old: {
            tags = [
              "postgres"
              "pgx"
              "pgx5"
            ];
          });

          devServices = pkgs.writeShellScriptBin "dev-services" ''
            set -euo pipefail

            # Anchor to the repo root so running from a subdirectory still uses
            # the same data dir. Data/logs live in ./.nix-dev; the top-level
            # .gitignore rule ("ignore everything unless allow-listed") keeps it
            # out of git.
            root="$(${pkgs.git}/bin/git rev-parse --show-toplevel 2>/dev/null || echo "$PWD")"
            devDir="$root/.nix-dev"

            # Overridable ports (avoid colliding with system Postgres/Redis).
            pgPort="''${PG_PORT:-5432}"
            valkeyPort="''${VALKEY_PORT:-6379}"
            devDbName="receipt_processor"

            pgData="$devDir/postgres"
            logDir="$devDir/log"
            # Unix socket paths are limited to ~107 bytes, so a deep checkout
            # could break a project-local socket dir. Nothing connects over the
            # socket (app + psql use TCP), so keep it short and per-user:
            # $TMPDIR if set, else /tmp.
            socketDir="''${TMPDIR:-/tmp}/rp-go-dev-''$(id -u)"
            valkeyDir="$devDir/valkey"

            case "''${1:-}" in
              start_pg)
                ${pkgs.postgresql_18}/bin/pg_ctl -D "$pgData" status >/dev/null 2>&1 \
                  && { echo "postgres: already running on port $pgPort"; exit 0; }

                if [ ! -e "$pgData/PG_VERSION" ]; then
                  echo "postgres: initializing data directory $pgData -- first run"
                  mkdir -p "$pgData"
                  # Trust auth (single-user dev cluster): connections with any
                  # password are accepted, so the app's default DSN with
                  # "postgres:postgres" works unchanged.
                  ${pkgs.postgresql_18}/bin/initdb -D "$pgData" -U postgres --auth=trust --no-locale --encoding=UTF8
                fi

                mkdir -p "$logDir" "$socketDir"
                ${pkgs.postgresql_18}/bin/pg_ctl -D "$pgData" \
                  -l "$logDir/postgres.log" \
                  -o "-p $pgPort -c listen_addresses=127.0.0.1 -c unix_socket_directories=$socketDir" \
                  -w start

                ${pkgs.postgresql_18}/bin/psql -h 127.0.0.1 -p "$pgPort" -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$devDbName'" | grep -q 1 \
                  || ${pkgs.postgresql_18}/bin/createdb -h 127.0.0.1 -p "$pgPort" -U postgres "$devDbName"
                echo "postgres: ready on 127.0.0.1:$pgPort (db: $devDbName)"
                ;;

              start_valkey)
                if ${pkgs.valkey}/bin/valkey-cli -h 127.0.0.1 -p "$valkeyPort" ping >/dev/null 2>&1; then
                  echo "valkey: already running on port $valkeyPort"
                else
                  mkdir -p "$valkeyDir" "$logDir"
                  ${pkgs.valkey}/bin/valkey-server \
                    --daemonize yes \
                    --bind 127.0.0.1 \
                    --port "$valkeyPort" \
                    --dir "$valkeyDir" \
                    --pidfile "$valkeyDir/valkey.pid" \
                    --logfile "$logDir/valkey.log"
                  # daemonize returns before the server listens; wait until it
                  # answers PING so `dev-services start && ... (something else)`
                  # actually yields ready services.
                  i=0
                  until ${pkgs.valkey}/bin/valkey-cli -h 127.0.0.1 -p "$valkeyPort" ping >/dev/null 2>&1; do
                    i=$((i + 1))
                    [ "$i" -lt 100 ] || { echo "valkey: failed to become ready" >&2; exit 1; }
                    sleep 0.1
                  done
                  echo "valkey: ready on 127.0.0.1:$valkeyPort"
                fi
                ;;

              start)
                "$0" start_pg
                "$0" start_valkey
                ;;

              stop)
                if ${pkgs.valkey}/bin/valkey-cli -h 127.0.0.1 -p "$valkeyPort" ping >/dev/null 2>&1; then
                  ${pkgs.valkey}/bin/valkey-cli -h 127.0.0.1 -p "$valkeyPort" shutdown && echo "valkey: stopped"
                else
                  echo "valkey: not running"
                fi

                if ${pkgs.postgresql_18}/bin/pg_ctl -D "$pgData" status >/dev/null 2>&1; then
                  ${pkgs.postgresql_18}/bin/pg_ctl -D "$pgData" stop -m fast && echo "postgres: stopped"
                else
                  echo "postgres: not running"
                fi
                ;;

              status)
                if ${pkgs.postgresql_18}/bin/pg_ctl -D "$pgData" status >/dev/null 2>&1; then
                  echo "postgres: running (127.0.0.1:$pgPort)"
                else
                  echo "postgres: not running"
                fi

                if ${pkgs.valkey}/bin/valkey-cli -h 127.0.0.1 -p "$valkeyPort" ping >/dev/null 2>&1; then
                  echo "valkey: running (127.0.0.1:$valkeyPort)"
                else
                  echo "valkey: not running"
                fi
                ;;

              *)
                echo "usage: dev-services {start|stop|status|start_pg|start_valkey}" >&2
                exit 1
                ;;
            esac
          '';
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go_1_27 # 1.27.x — go.mod targets Go 1.27
              gopls # go language server for editors
              delve # go debugger
              just # build/tasks runner (justfile)
              git # used by `just version` (git describe) and everyday work
              goMigrate # golang-migrate CLI (only postgres/pgx drivers, see above)
              nixfmt # Nix formatter (RFC 166 style) for editor integration
              golangci-lint # go linter aggregator (run via `just lint`)
              postgresql_18 # psql + server — run via `dev-services start`
              valkey # BSD-licensed redis-compatible server — run via `dev-services start`
              devServices
            ];

            shellHook = ''
              # The app defaults to this DSN, but the justfile's migrate recipes
              # read $DSN directly — make it available there too. Set in .env to
              # override.
              export DSN="''${DSN:-postgres://postgres:postgres@localhost:5432/receipt_processor?sslmode=disable}"

              echo "dev shell: go + postgres/valkey ready — run \`dev-services start\`"
            '';
          };
        }
      );
    };
}
