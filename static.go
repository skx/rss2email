//
// This file was generated via github.com/skx/implant/
//
// Local edits will be lost.
//
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
)

//
// EmbeddedResource is the structure which is used to record details of
// each embedded resource in your binary.
//
// The resource contains the (original) filename, relative to the input
// directory `implant` was generated with, along with the original size
// and the compressed/encoded data.
//
type EmbeddedResource struct {
	Filename string
	Contents string
	Length   int
}

//
// RESOURCES is a map containing all embedded resources. The map key is the
// file name.
//
// It is exposed to callers via the `getResources()` function.
//
var RESOURCES = map[string]EmbeddedResource{

	"data/email.tmpl": {
		Filename: "data/email.tmpl",
		Contents: "H4sIAAAAAAAC/6yTz47TMBDG736KaO/e+l9jp9v2AlQcdi+0ICTEYWxPaCBxguOirqq8O0oXCqtVT9nb+PPM6DefZt60IWFIdPfY4SJrDnWqOohp1lRH9HeZbQ/BQ3xcCY4oPRS5ctqWmnkscwHeFogcWM610oxJx0SOMpeotJBczI1TpuRkE9tmkZ1Ot2MwDGTXnl+7dhjI9mC/o0uL7Evse4ENVPXX8fOPPgzkM/2w3dL7Kvw4V43BRd0g+qfOiH4YyEPVIP2Esa/asMj4LSOE0kns1wyKWEN6ZpHOgQsjpJaOGcjn3qvCW+4RhJvPOeelhwI0nztTaMHAKFWCtdqrQo+Yk+qvYUKdMAZI1S/8D1VxkztZIJdWcMWcMaxQVhZSCpZLmZfCWlko461GV7LRONBeKy1LMaJOqn+OmvCYZl0NVbjL3B5ij2n1cbeh5l9ehNCXGOm74FpfhW+L7OehTehpF6uQwNZIyOl01i5K9ndPyLhoeExP4ZWk159on5p60kDLbr2EbB+xXMm3N9fIb9Yvfy6Xs5zBejnr1qMF73cP98Pwql2nmUbp1J2ndOJtU0p+BwAA//+Y393l/wQAAA==",
		Length:   1279,
	},
}

//
// Return the contents of a resource.
//
func getResource(path string) ([]byte, error) {
	if entry, ok := RESOURCES[path]; ok {
		var raw bytes.Buffer
		var err error

		// Decode the data.
		in, err := base64.StdEncoding.DecodeString(entry.Contents)
		if err != nil {
			return nil, err
		}

		// Gunzip the data to the client
		gr, err := gzip.NewReader(bytes.NewBuffer(in))
		if err != nil {
			return nil, err
		}
		defer gr.Close()
		data, err := ioutil.ReadAll(gr)
		if err != nil {
			return nil, err
		}
		_, err = raw.Write(data)
		if err != nil {
			return nil, err
		}

		// Return it.
		return raw.Bytes(), nil
	}
	return nil, fmt.Errorf("failed to find resource '%s'", path)
}

//
// Return the available resources in a slice.
//
func getResources() []EmbeddedResource {
	i := 0
	ret := make([]EmbeddedResource, len(RESOURCES))
	for _, v := range RESOURCES {
		ret[i] = v
		i++
	}
	return ret
}
