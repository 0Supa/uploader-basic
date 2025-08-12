{
  description = "Basic upload server";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      module = "uploader-basic";
      user = "uploader-basic";
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
      goServer = pkgs.buildGoModule {
        pname = "uploader-basic";
        version = "git-${self.rev or "dirty"}";
        src = ./.;
        vendorHash = null;
      };
    in
    {
      packages.${system}.default = goServer;

      nixosModules.${module} =
        { config, pkgs, ... }:
        {
          users.users.${user} = {
            isSystemUser = true;
            home = "/var/lib/${user}";
            group = "www";
          };

          systemd.services.${module} = {
            description = "Simple Go HTTP server";
            after = [ "network.target" ];
            wantedBy = [ "multi-user.target" ];
            serviceConfig = {
              ExecStart = "${goServer}/bin/${module}";
              Restart = "always";
              User = user;
              WorkingDirectory = "/var/lib/${user}";
            };
          };
        };
    };
}
