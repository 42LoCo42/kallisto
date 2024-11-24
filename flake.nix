{
  inputs.nix-appimage = {
    url = "github:42LoCo42/nix-appimage";
    inputs = {
      flake-utils.follows = "flake-utils";
      nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { flake-utils, nixpkgs, nix-appimage, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; }; in rec {
        packages = {
          default = pkgs.buildGoModule rec {
            pname = "kallisto";
            version = "1";

            src = with pkgs.lib.fileset; toSource {
              root = ./.;
              fileset = unions [
                ./go.mod
                ./go.sum
                ./main.go
              ];
            };

            nativeBuildInputs = with pkgs; [ cacert ];
            buildInputs = with pkgs; [ libsodium ];
            ldflags = [ "-s" "-w" ];
            vendorHash = "sha256-6Qy5AN0cH6TJOspuPYUYQjGjFJwchBGg1ngS/4QCGC4=";

            meta.mainProgram = pname;
          };

          appimage = nix-appimage.bundlers.${system}.default packages.default;
        };

        devShells.default = pkgs.mkShell {
          inputsFrom = [ packages.default ];
        };
      });
}
