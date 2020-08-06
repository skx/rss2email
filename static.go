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
		Contents: "H4sIAAAAAAAC/6xUTW/jNhC961cM9rgAbX1ZH97EQLFt2kP20LVSFCh6GJEji12JUkkqjhH4vxeUv5Tt+uTwZHDmPT2+eePX1/lHKGppQBqwNYGgCofGgqW2b9ASbGvJa1cdDAkod5eODjakSLsmalE2ZuZ54M5PBnbdAK3c1BZkixupCKR1JAi/dg2qDVh6sfPzRyrZ0Am9pmfS2EAlqREGUAmoBsWt7JQB1AT4jLLBsqHlEQGvr7MHIlFI29B+D8CgqAnqoUXFNKFwzWBdFbpqfKbpBs0JKiIxe0vi8OM5kDx9fTxhXDNUumuPnrg7aakFji1NWXTXfscy+gMohCZjjnBDShw8H6sTgkepvn1H0Ej1zTnu2hVtgZTVuwlkPZT/ELcH1AFiDlcn9T9CFd35MyeUJi57SeqMO4o7gB5Og5hY/+/QWeq1VHa0+aKereB3VzsY3hOXlXQOurleRKhuOxExOWwF67rbjmg+aO00CbQ0t7KlGdBsMzvAZ0/F5/3+KGjMMu/a1rVvZdNASWCsln1/mp4jPCVXTN73cb7ff+6UJWVZsetpCe3QWNmjtvNWvpD4BGU3KIF6dx8GRJHAPIl5WlapL6hKQhRlThSgnwRpnPp+xP0woSiJKE7DKAgXGY+zKvBcQJaXpHhFtzxNwzsOcgl/aWPCUdzfbwbs/cm+rtfMuby8hOV46wK8vCTZ+yJbYn+QNrJTSwhmvucxdpP2awZpcos8tShNMAizMEoj7meYLISIc1EGgjDki0UQBJXAHNNgwbM8DX3M4rjCskxFnKdO5k34azKxsaQVWvlME6lxkCU8yimIyjCIfZ5lfh6XUR5FoZ9EUVKFZRnlcSbKlHjlO+MwFWmcRlXopN6Efyt1/F/sG5TqE/AatSF7/1Q8sOzSp1GZijT7RfFOSLVZwriCgp2X0POuraWrzAp6sYefV5re/0W1bZubHnTXr+4Qak3VffTzh2vKP6z+Xzlvzt0cV3fzfuUs+K348rjfvyvrbaYxdmvmGbtxtxnz/gsAAP//WyM5qQ8IAAA=",
		Length:   2063,
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
