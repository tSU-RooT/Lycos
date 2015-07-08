package lycos

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/http"
)

func serialize(item interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	encode := gob.NewEncoder(buf)
	if err := encode.Encode(item); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserialize(value []byte, item interface{}) error {
	decode := gob.NewDecoder(bytes.NewBuffer(value))
	if err := decode.Decode(item); err != nil {
		return err
	}
	return nil
}

func memcacheKey(k string, properties ...interface{}) string {
	v := k + ":"
	for i := range properties {
		v += fmt.Sprintf(" %v", properties[i])
	}
	return v
}

func bad(w http.ResponseWriter) {
	if err := badTmpl.ExecuteTemplate(w, "base", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
