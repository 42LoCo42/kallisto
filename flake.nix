{
  outputs = { flake-utils, nixpkgs, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; }; in rec {
        packages.default = pkgs.buildGoModule {
          pname = "kallisto";
          version = "1";
          src = ./.;
          vendorHash = "sha256-upIQ9mIuTR0ruiJ6gyJfh+10Ahuic4wixpnbJz5LES4=";

          buildInputs = with pkgs; [ libsodium ];

          ldflags = [ "-s" "-w" ];
        };

        devShells.default = pkgs.mkShell {
          inputsFrom = [ packages.default ];
        };
      });
}
