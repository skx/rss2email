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
		Contents: "H4sIAAAAAAAC/6xUT2/buBO961MMeiwqW/9iyW4S4IdfN90C6WFjZbHAYg8jchRxS5FaksofGP7uC0qW5bSbk6uTQM57evPmjXa75XsoG2FBWHANAacae+nAUdtJdARPjWCNv+0tcahe5goND6TI+CJqUUi7CALwz/8svOgeWvHQOBAtPghFIJwnQfisJaoHcPTslseP1ELShN7SIxmUUAuS3AIqDnWvmBNaWUBDgI8oJFaSNgcE7HaLGyJeCidpvwcIoWwImr5FFRpC7ovB+VvQ9dCm1b1hBDURX7wm8fjhGUnu724njC+G2uj24Ik/E45aYNjSKYvR7Xcsgz+AnBuy9gC3pPjo+XB7QnAr1LfvCKRQ37zjvlzRE5By5uUEsu2rv4m5ETVC7Hg0qf8vVKmPn5lQhpjoBKkj7iBuBJUNGRrmKK0GZMy3c5C12y3uttuDhX5q48EXR96N49A+gFAjGUNLQ1CUN3Ym064hM41/9MqPXWkH9Nxp44j7FymYdFM391b4UDXoALvOaGTNQM1QTcyznMXn+y+f9vsPUGsD9IxtN6fvZoraSbj+6bWjzgjlhiDN8wmv4Td/N0aqIyZq4TPipc+WCQtMt6339ElICRWBdUZ03RQnj55WiZ8Y/n653/9fK0fKheVLRxtoe+lEh8YtW/FM/CNUulcczctVEhOlHNerjOVVnUec6lWCvFoTxRit4jzLoyhlUbKidJVSlidpnFwULCvqOPCJ3czRDUq9meIRHJK1gT+Ntckg7q9XiQv+CO+229CbspnTezj1cdjMq3U49f5vfhxI8FW0FP5OxgqtNhAvoiAIw7M6e8s+Q/6/c2pgvsI4KZI0T1lU4OqC82zNq5gTJuziIo7jmuMa8/iCFes8ibDIshqrKufZOvcyz8K/JROlI6PQiUc6kZrFxYqla4rTKomziBVFtM6qdJ2mSbRK01WdVFW6zgpe5cTqyBuHOc+zPK0TL/Us/Gupw2+8kyjUR2ANGkvu6r68CYu5zqCyNZnwF8U0F+phA8M+8fC4UUHw1o75m0VJz258faPo53fUuFae1dBld32J0Biqr9JP795S/u76x5vjXl0u8fpy2V17C34tv97u9z+V9TzTwvDczIfhmbsdhsG/AQAA///oIsFdvggAAA==",
		Length:   2238,
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
