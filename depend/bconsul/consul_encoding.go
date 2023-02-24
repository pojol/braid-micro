package bconsul

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
)

func encode(buf []byte) string {
	var b bytes.Buffer
	defer b.Reset()

	w := zlib.NewWriter(&b)
	if _, err := w.Write(buf); err != nil {
		return ""
	}
	w.Close()

	return hex.EncodeToString(b.Bytes())
}

func decode(d string) []byte {
	hr, err := hex.DecodeString(d)
	if err != nil {
		return nil
	}

	br := bytes.NewReader(hr)
	zr, err := zlib.NewReader(br)
	if err != nil {
		return nil
	}

	rbuf, err := ioutil.ReadAll(zr)
	if err != nil {
		return nil
	}
	zr.Close()

	return rbuf
}

func encodeMetadata(md map[string]string) []string {
	var tags []string
	for k, v := range md {
		if b, err := json.Marshal(map[string]string{
			k: v,
		}); err == nil {
			// new encoding
			tags = append(tags, "t-"+encode(b))
		}
	}
	return tags
}

func decodeMetadata(tags []string) map[string]string {
	md := make(map[string]string)

	var ver byte

	for _, tag := range tags {
		if len(tag) == 0 || tag[0] != 't' {
			continue
		}

		// check version
		if ver > 0 && tag[1] != ver {
			continue
		}

		var kv map[string]string
		var buf []byte

		// Old encoding was plain
		if tag[1] == '=' {
			buf = []byte(tag[2:])
		}

		// New encoding is hex
		if tag[1] == '-' {
			buf = decode(tag[2:])
		}

		// Now unmarshal
		if err := json.Unmarshal(buf, &kv); err == nil {
			for k, v := range kv {
				md[k] = v
			}
		}

		// set version
		ver = tag[1]
	}
	return md
}
