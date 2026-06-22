package main

import (
	"BallexSeriesCustomMapDataAnalyzer/ballex2"
	"BallexSeriesCustomMapDataAnalyzer/bme4"
	"BallexSeriesCustomMapDataAnalyzer/util"
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/tidwall/gjson"
)

const DIR = `D:\Steam\steamapps\workshop\content\1383570\3748466204`

func main() {
	entries, _ := os.ReadDir(DIR)
	if len(entries) == 3 {
		if entries[0].Name() == `BallexMap.bms` && entries[1].Name() == `BallexMapCover.jpg` && entries[2].Name() == `WorkshopItemInfo.xml` {
			raw, _ := os.ReadFile(DIR + `/BallexMap.bms`)
			key, _ := pbkdf2.Key(sha1.New, `BallexFilePasswordIsWhat?TheAnswerIsIDoNotKnow!`, raw[:16], 100, 16)
			block, _ := aes.NewCipher(key)
			blockMode := cipher.NewCBCDecrypter(block, raw[:16])
			decry := raw[16:]
			for i := range len(decry) >> 4 {
				blockMode.CryptBlocks(decry[i<<4:i<<4+16], decry[i<<4:i<<4+16])
			}
			leng := len(decry)
			unpadding := int(decry[leng-1])
			if unpadding > 16 {
				fmt.Println(`unpadding不正常`, unpadding)
			}
			decry = slices.Delete(decry, leng-unpadding, leng)

			os.WriteFile(`./b1scene.json`, decry, 0644)
			names := slices.DeleteFunc(gjson.GetBytes(raw, `ExtrasName.value`).Array(), func(r gjson.Result) bool { return r.Str != `ScoreBall` })
			fmt.Println(len(names), `个分数球`)
		} else if entries[0].Name() == `Cover.jpg` && entries[1].Name() == `Upload.bms` && entries[2].Name() == `Upload.bms.assets` {
			ballex2.ReadBallex2MapData(DIR)
		}
	} else if len(entries) == 1 {
		raw, _ := os.ReadFile(DIR + `/` + entries[0].Name())
		block, _ := aes.NewCipher([]byte(`Ballex²AlphaTes`))
		blockMode := cipher.NewCBCDecrypter(block, raw[:16])
		decry := raw[16:]
		for i := range len(decry) >> 4 {
			blockMode.CryptBlocks(decry[i<<4:i<<4+16], decry[i<<4:i<<4+16])
		}
		leng := len(decry)
		unpadding := int(decry[leng-1])
		if unpadding > 16 {
			fmt.Println(`unpadding`, unpadding)
		}
		decry = slices.Delete(decry, leng-unpadding, leng)

		reader := bytes.NewReader(decry)
		reader.Seek(-4, io.SeekEnd)
		final := make([]byte, util.Read[uint32](reader))
		reader.Seek(0, io.SeekStart)

		gzipread, _ := gzip.NewReader(reader)
		io.ReadFull(gzipread, final)
		gzipread.Close()

		os.WriteFile(`./bme4.dat`, final, 0644)
		bme4.ReadBME4MapData(final)
	}
}
