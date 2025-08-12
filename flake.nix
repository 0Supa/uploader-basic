{
  description = "Basic upload server";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      pname = "uploader-basic";
      user = "uploader-basic";
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
      goServer = pkgs.buildGoModule {
        pname = pname;
        version = "git-${self.rev or "dirty"}";
        src = ./.;
        vendorHash = null;
      };
    in
    {
      packages.${system}.default = goServer;

      nixosModules.default =
        { config, pkgs, ... }:
        {
          users.users.${user} = {
            isSystemUser = true;
            home = "/var/lib/${user}";
            createHome = true;
            group = "www";
          };

          systemd.services.${pname} = {
            after = [ "network.target" ];
            wantedBy = [ "multi-user.target" ];
            serviceConfig = {
              ExecStart = "${goServer}/bin/${pname}";
              Restart = "always";
              User = "${user}";
              WorkingDirectory = "/var/lib/${user}";
            };
          };
        };
    };
}
