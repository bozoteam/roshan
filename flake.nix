{
  description = "Docker image for Go app and Node.js app";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    self.submodules = true;
  };
  outputs =
    inputs@{
      self,
      nixpkgs,
      flake-utils,
      ...
    }:

    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        buildTime = builtins.readFile (
          pkgs.runCommand "build-time" { } ''
            date +%s > $out
          ''
        );

        ship = pkgs.mkShell {
          packages = [
            pkgs.openssh
            pkgs.coreutils
          ];
          shellHook =
            let
              cat = "${pkgs.coreutils}/bin/cat";
              sshHost = "ec2-user@bozo.mateusbento.com";
              sshCommand = "${pkgs.openssh}/bin/ssh -i ~/.ssh/proxyaccess.pem ${sshHost}";
            in

            ''
              set -ex -o pipefail
              nix build .#packages.x86_64-linux.roshanDocker --out-link result-roshan

              ${sshCommand} 'cd /bozo && docker-compose down roshan'

              ${cat} result-roshan | ${sshCommand} 'docker rmi roshan -f ; docker load'

              ${sshCommand} 'cd /bozo && docker-compose up --build -d'

              echo SUCCESS
              exit 0
            '';
        };

        roshanApp = pkgs.buildGoModule {
          pname = "roshan";
          version = "0.1.0";
          src = ./.;
          subPackages = [ "cmd/server" ];
          vendorHash = "sha256-KyZmLUV5EDHEPOWiMyv78AmxGXcnZ7J4TmOomNvw1GM=";
          ldflags = [
            "-X github.com/bozoteam/roshan/helpers.development=false"
            "-X github.com/bozoteam/roshan/helpers.BuildTime=${buildTime}"
          ];
          env.CGO_ENABLED = 0;
        };
        roshanDockerImage =
          let
            dbUrl = "postgres://roshan:roshan@postgres:5432/roshan?sslmode=disable";
            migrationsDir = ./db/migrations;
            atlasConfig = pkgs.writeText "atlas.hcl" ''
              env "migrations" {
                url = "${dbUrl}"
                migration {
                  dir = "file:///migrations"
                  format = atlas
                 }
              }
            '';
          in
          pkgs.dockerTools.buildLayeredImage {
            name = "roshan";
            tag = "latest";
            contents = [
              roshanApp
              pkgs.atlas
              pkgs.bash
              pkgs.coreutils
            ];
            config = {
              Cmd = [
                "${pkgs.bash}/bin/bash"
                "-c"
                "${pkgs.atlas}/bin/atlas migrate apply --env migrations && ${roshanApp}/bin/server"
              ];
              Env = [
                "DATABASE_URL=\"${dbUrl}\""
              ];
              ExposedPorts = {
                "8080/tcp" = { };
              };
            };
            extraCommands = ''
              mkdir -p migrations
              cp -r ${migrationsDir}/* migrations/
              cp ${atlasConfig} atlas.hcl
            '';
          };
      in
      {
        devShells = {
          ship = ship;
        };
        packages = {
          default = roshanDockerImage;
          roshan = roshanApp;
          roshanDocker = roshanDockerImage;
        };
      }
    );
}
