{
  description = "Docker image for Go app and Node.js app";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    self.submodules = true;
  };
  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      ...
    }:

    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        ship = pkgs.mkShell {
          packages = [
            pkgs.openssh
            pkgs.coreutils
          ];
          shellHook =
            let
              cat = "${pkgs.coreutils}/bin/cat";
              sshCommand = "${pkgs.openssh}/bin/ssh -i ~/.ssh/proxyaccess.pem ec2-user@bozo.mateusbento.com";
            in
            ''
              set -ex -o pipefail
              nix build .#packages.x86_64-linux.roshanDocker --out-link result-roshan
              nix build .#packages.x86_64-linux.atlasDocker --out-link result-atlas

              ${sshCommand} 'cd /bozo && docker-compose down roshan'

              ${cat} result-atlas  | ${sshCommand} 'docker rmi atlas  -f ; docker load'

              ${sshCommand} 'cd /bozo && docker run --network=bozo_backend atlas'

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
          vendorHash = "sha256-oZf1IpHyQBMtETIc6J6WJu1XwzLk28HHMBm9QNvQ0/g=";
        };
        # More minimal Docker image using a stripped binary
        roshanDockerImage = pkgs.dockerTools.buildLayeredImage {
          name = "roshan";
          tag = "latest";
          contents = [ roshanApp ];
          config = {
            Cmd = [ "${roshanApp}/bin/server" ];
            ExposedPorts = {
              "8080/tcp" = { };
            };
          };
        };

        atlasDockerImage =
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
            name = "atlas";
            tag = "latest";
            contents = [
              pkgs.atlas
              pkgs.bash
              pkgs.coreutils
            ];
            config = {
              Cmd = [
                "${pkgs.bash}/bin/bash"
                "-c"
                "${pkgs.atlas}/bin/atlas migrate apply --env migrations"
              ];
              Env = [
                "DATABASE_URL=\"${dbUrl}\""
              ];
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
          atlasDocker = atlasDockerImage;
        };
      }
    );
}
