package assets

import "github.com/thrisp/flotilla/resources"

var TestAsset AssetFS = NewAssetFS(
	resources.Asset,
	resources.AssetDir,
	resources.AssetNames,
	"",
)

/*
func TestAssets(t *testing.T) {
	//spew.Dump(TestAsset)
	//f, _ := resources.Asset("/assets/templates/test_asset.html")
	//fd, _ := resources.AssetDir("assets")
	//names := resources.AssetNames()
	//fl, _ := TestAsset.Open(fmt.Sprintf("%s", f))
	//fmt.Printf("%+v\n", checkff)
}
*/
