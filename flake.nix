{
  description = "Go 1.27.0 development environment";

  inputs = {
    # go_1_27 = 1.27.0 を含むリビジョンにピン留め
    nixpkgs.url = "github:NixOS/nixpkgs/e8be7818e19ada32105a8af937a6a473b38167ca";
  };

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = [
            pkgs.go_1_27
            pkgs.gopls
            pkgs.gotools
            pkgs.go-tools
            pkgs.delve
            pkgs.air
          ];

          shellHook = ''
            export GOROOT="${pkgs.go_1_27}/share/go"
            export GOPATH="$PWD/.go"
            export PATH="$GOPATH/bin:$PATH"
            echo "$(go version)"
          '';
        };
      });
    };
}
