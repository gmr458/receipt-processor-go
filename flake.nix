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
      formatter = forAllSystems (pkgs: pkgs.nixfmt-tree);

      devShells = forAllSystems (
        pkgs:
        let
          goMigrate = pkgs.go-migrate.overrideAttrs (old: {
            tags = [
              "postgres"
              "pgx"
              "pgx5"
            ];
          });

          delve = pkgs.delve.override {
            buildGoModule = pkgs.buildGoModule.override { go = pkgs.go_1_27; };
          };

          devServices = pkgs.writeShellScriptBin "dev-services" ''
            set -euo pipefail

            root="$(${pkgs.git}/bin/git rev-parse --show-toplevel 2>/dev/null || echo "$PWD")"
            devDir="$root/.nix-dev"

            pgPort="''${PG_PORT:-5432}"
            valkeyPort="''${VALKEY_PORT:-6379}"
            devDbName="receipt_processor"

            pgData="$devDir/postgres"
            logDir="$devDir/log"
            socketDir="''${TMPDIR:-/tmp}/rp-go-dev-''$(id -u)"
            valkeyDir="$devDir/valkey"

            case "''${1:-}" in
              start_pg)
                ${pkgs.postgresql_18}/bin/pg_ctl -D "$pgData" status >/dev/null 2>&1 \
                  && { echo "postgres: already running on port $pgPort"; exit 0; }

                if [ ! -e "$pgData/PG_VERSION" ]; then
                  echo "postgres: initializing data directory $pgData -- first run"
                  mkdir -p "$pgData"
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
              go_1_27
              gopls
              delve
              just
              git
              goMigrate
              nixfmt
              golangci-lint
              postgresql_18
              valkey
              devServices
            ];

            shellHook = ''
              export DSN="''${DSN:-postgres://postgres:postgres@localhost:5432/receipt_processor?sslmode=disable}"

              echo "dev shell: go + postgres/valkey ready — run \`dev-services start\`"
            '';
          };
        }
      );
    };
}
