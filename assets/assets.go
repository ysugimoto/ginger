package assets

import (
	"time"

	"github.com/jessevdk/go-assets"
)

var _Assetsdbfa9bb4fc5340440719a646c285441fc9134ddd = "SG9tZSBNYWRlIEdpbmdlciBBbGUKCiMjLSBJTkdSRURJRU5UUwotIDEgMS8yIGN1cHMgY2hvcHBlZCBwZWVsZWQgZ2luZ2VyICg3IG91bmNlcykKLSAyIGN1cHMgd2F0ZXIKLSAzLzQgY3VwIHN1Z2FyCi0gQWJvdXQgMSBxdWFydCBjaGlsbGVkIHNlbHR6ZXIgb3IgY2x1YiBzb2RhCi0gQWJvdXQgMyB0YWJsZXNwb29ucyBmcmVzaCBsaW1lIGp1aWNlCgojIyBQUkVQQVJBVElPTgoKIyMjIE1ha2Ugc3lydXA6CkNvb2sgZ2luZ2VyIGluIHdhdGVyIGluIGEgc21hbGwgc2F1Y2VwYW4gYXQgYSBsb3cgc2ltbWVyLCBwYXJ0aWFsbHkgY292ZXJlZCwgNDUgbWludXRlcy4gUmVtb3ZlIGZyb20gaGVhdCBhbmQgbGV0IHN0ZWVwLCBjb3ZlcmVkLCAyMCBtaW51dGVzLgpTdHJhaW4gbWl4dHVyZSB0aHJvdWdoIGEgc2lldmUgaW50byBhIGJvd2wsIHByZXNzaW5nIG9uIGdpbmdlciBhbmQgdGhlbiBkaXNjYXJkaW5nLiBSZXR1cm4gbGlxdWlkIHRvIHNhdWNlcGFuIGFuZCBhZGQgc3VnYXIgYW5kIGEgcGluY2ggb2Ygc2FsdCwgdGhlbiBoZWF0IG92ZXIgbWVkaXVtIGhlYXQsIHN0aXJyaW5nLCB1bnRpbCBzdWdhciBoYXMgZGlzc29sdmVkLiBDaGlsbCBzeXJ1cCBpbiBhIGNvdmVyZWQgamFyIHVudGlsIGNvbGQuCgojIyMgQXNzZW1ibGUgZHJpbmtzOgpNaXggZ2luZ2VyIHN5cnVwIHdpdGggc2VsdHplciBhbmQgbGltZSBqdWljZSAoc3RhcnQgd2l0aCAxLzQgY3VwIHN5cnVwIGFuZCAxIDEvMiB0ZWFzcG9vbnMgbGltZSBqdWljZSBwZXIgMy80IGN1cCBzZWx0emVyLCB0aGVuIGFkanVzdCB0byB0YXN0ZSkuCg==\n"
var _Assets7e9bedb64677d9d3d7167b59bc342e1eb5dbd8d8 = "package main\n\nimport (\n\t\"context\"\n\n\t\"github.com/aws/aws-lambda-go/lambda\"%s\n)\n\nfunc %sHandler(ctx context.Context, %s) %s {\n\treturn %s\n}\n\nfunc main() {\n\tlambda.Start(%sHandler)\n}\n"

// Assets returns go-assets FileSystem
var Assets = assets.NewFileSystem(map[string][]string{"/": []string{"ale", "main.go.template"}}, map[string]*assets.File{
	"/main.go.template": &assets.File{
		Path:     "/main.go.template",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1519130263, 1519130263000000000),
		Data:     []byte(_Assets7e9bedb64677d9d3d7167b59bc342e1eb5dbd8d8),
	}, "/": &assets.File{
		Path:     "/",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1519130263, 1519130263000000000),
		Data:     nil,
	}, "/ale": &assets.File{
		Path:     "/ale",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1518347191, 1518347191000000000),
		Data:     []byte(_Assetsdbfa9bb4fc5340440719a646c285441fc9134ddd),
	}}, "")
