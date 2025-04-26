{
  description = "Application with Go, Atlas, and PostgreSQL";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        # Main run script that starts everything
        runScript = pkgs.writeScriptBin "run-app" ''
          #!${pkgs.fish}/bin/fish

          # Setup cleanup function
          function cleanup
            echo "Cleaning up..."
            ${pkgs.postgresql}/bin/pg_ctl -D $PGDATA stop
            pkill -P (echo %self)
            exit 0
          end

          # Setup trap for cleanup
          trap cleanup SIGINT SIGTERM

          # Setup PostgreSQL
          set -x PGDATA "$PWD/postgres_data"

          # Initialize PostgreSQL if needed
          if test ! -d "$PGDATA"
            echo "Initializing PostgreSQL database..."
            mkdir -p "$PGDATA"

          # Initialize with explicit postgres superuser
          ${pkgs.postgresql}/bin/initdb -D "$PGDATA" \
            --username=postgres \
            --auth=trust \
            --no-locale \
            --encoding=UTF8

            # Configure PostgreSQL
            echo "listen_addresses = '127.0.0.1'" >> "$PGDATA/postgresql.conf"
            echo "port = 5432" >> "$PGDATA/postgresql.conf"
            
            # Set up authentication
            echo "local all all trust" > "$PGDATA/pg_hba.conf"
            echo "host all all 127.0.0.1/32 trust" >> "$PGDATA/pg_hba.conf"
            echo "host all all ::1/128 trust" >> "$PGDATA/pg_hba.conf"
            
            # Start PostgreSQL
            ${pkgs.postgresql}/bin/pg_ctl -D "$PGDATA" -l "$PGDATA/logfile" start

          else
            # Just start the server
            ${pkgs.postgresql}/bin/pg_ctl -D "$PGDATA" -l "$PGDATA/logfile" start
          end

          # Wait for PostgreSQL to be ready
          echo "Waiting for PostgreSQL to be ready..."
          while true
            ${pkgs.postgresql}/bin/pg_isready -U postgres -h 127.0.0.1
            if test $status -eq 0
              echo "PostgreSQL is ready!"
              break
            end
            sleep 1
          end

          # Create user
          ${pkgs.postgresql}/bin/createuser roshan -d -h localhost -U postgres
          # Create database
          ${pkgs.postgresql}/bin/createdb roshan -O roshan -U roshan

          # Run migrations and schema updates
          ${pkgs.atlas}/bin/atlas migrate diff $argv --env postgres --config "file://db/atlas.hcl"; or begin; cleanup; exit 1; end
          ${pkgs.atlas}/bin/atlas migrate apply --env postgres --config file://db/atlas.hcl; or begin; cleanup; exit 1; end

          # Run the Go application
          echo "Starting application..."
          cd $PWD && ${pkgs.go}/bin/go run ./cmd/server/main.go; or begin; cleanup; exit 1; end
        '';

      in
      {
        packages = {
          run-app = runScript;
          default = runScript;
        };

        devShell = pkgs.mkShell {
          buildInputs = [
            pkgs.go
            pkgs.atlas
            pkgs.postgresql
            pkgs.fish
            runScript
          ];

          shellHook = ''
            echo "Development environment ready!"
            echo "Run 'run-app' to start the application with PostgreSQL"
            echo "Or use individual commands: 'migrate', 'apply'"
          '';
        };
      }
    );
}
