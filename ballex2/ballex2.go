package ballex2

import (
	"BallexSeriesCustomMapDataAnalyzer/util"
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode/utf16"
)

type BmsAssetPack struct {
	GUID                 string
	AssetsPackageVersion int32
	Assets               map[string]BmsAsset
}

type BmsAsset struct {
	AssetType        uint64
	IsBuiltInAsset   bool
	BuiltInAssetLink string
	AssetData        any
}

type MushTextureAsset struct {
	Width, Height int32
	Data          []byte
	FilterMode    uint64
	MipChain, HDR bool
	Linear        bool
}

type MushMeshAsset struct {
	Vertices           [][3]float32
	SubMeshDescriptors []MushMeshDescriptor
	UVs                [][2]float32
	Normals            [][3]float32
	Tangents           [][4]float32
}

type MushMeshDescriptor struct {
	IndexStart, IndexCount    int32
	SubMeshTriangles, Indices []int32
}

type MushMaterialAsset struct {
	MaterialType                   int32
	Albedo, Emission, Normal, Mask string
	AlbedoColor, EmissionColor     [4]float32
	EmissionIntensity, EmissionIntensityMin, EmissionTwinkleSpeed,
	Metallic, Smoothness, AO, NormalScale float32
	OverlayTrack bool
	OverlayNoiseScale, OverlayNoiseBlend, OverlayClamp,
	OverlayProjection, OverlayOffset, WetnessClamp,
	WetnessOffset, WetnessSmoothnessMultiplier float32
	TilingScale, TilingOffset, TilingSpeed [2]float32
	GlobalUV                               bool
	GlobalUVTile, GlobalUVBlend            float32
	TopAlbedo, TopNormal, TopMask          string
	TransparencyType                       int32
	TransparencyDither                     bool
	AlphaClipThreshold                     float32
	AffectAlbedo, AffectNormal, AffectMetal, AffectAO,
	AffectSmoothness, AffectEmission bool
	DrawOrder int32
}

func ReadBallex2MapData(path string) {
	raw, _ := os.ReadFile(path + `/Upload.bms`)
	reader := bytes.NewReader(raw)
	reader.Seek(168, io.SeekStart)
	xmlBytes := make([]byte, util.Read[int32](reader))
	reader.Seek(4, io.SeekCurrent)
	binary.Read(reader, binary.LittleEndian, &xmlBytes)
	tokens := decodeBinaryXML(xmlBytes)
	if true {
		f, _ := os.Create(`./b2scene.xml`)
		encoder := xml.NewEncoder(f)
		encoder.Indent(``, "\t")
		for _, token := range tokens {
			encoder.EncodeToken(token)
		}
		encoder.Flush()
		encoder.Close()
		f.Close()
	}

	pack := readMapAssets(path)
	if true {
		os.Mkdir(`./export`, 0644)
		for k, v := range pack.Assets {
			switch v.AssetType {
			case 2:
				os.WriteFile(`./export/`+strings.NewReplacer(`/`, `.`, `.tex`, `.png`).Replace(k), v.AssetData.(MushTextureAsset).Data, 0644)
			case 5:
				os.WriteFile(`./export/`+strings.NewReplacer(`/`, `.`, `.audio`, `.ogg`).Replace(k), v.AssetData.([]byte), 0644)
			case 6:
				os.WriteFile(`./export/`+strings.NewReplacer(`/`, `.`).Replace(k), []byte(v.AssetData.(string)), 0644)
			}
		}
	}
}

func readMapAssets(path string) (pack BmsAssetPack) {
	raw, _ := os.ReadFile(path + `/Upload.bms.assets`)
	reader := bytes.NewReader(raw)
	reader.Seek(111, io.SeekStart)
	pack.GUID = util.ReadString(reader)
	reader.Seek(46, io.SeekCurrent)
	binary.Read(reader, binary.LittleEndian, &pack.AssetsPackageVersion)

	typeCache := map[int32]string{}
	readType := func() {
		typ, _ := reader.ReadByte()
		switch typ {
		case 47:
			index := util.Read[int32](reader)
			typeCache[index] = util.ReadString(reader)
		case 48:
			util.Read[int32](reader)
		}
	}

	reader.Seek(420, io.SeekCurrent)
	assetCount := util.Read[uint64](reader)
	pack.Assets = make(map[string]BmsAsset, assetCount)
	i := 0
	for range assetCount {
		asset := BmsAsset{}
		fmt.Println(`资源`, i, `位于`, reader.Size()-int64(reader.Len()))

		reader.Seek(12, io.SeekCurrent)
		key := util.ReadString(reader)
		reader.Seek(10, io.SeekCurrent)
		readType()
		reader.Seek(28, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &asset.AssetType)
		reader.Seek(34, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &asset.IsBuiltInAsset)
		reader.Seek(38, io.SeekCurrent)
		asset.BuiltInAssetLink = util.ReadString(reader)
		if asset.IsBuiltInAsset {
			reader.Seek(26, io.SeekCurrent)
			pack.Assets[key] = asset
			i++
			continue
		}

		reader.Seek(24, io.SeekCurrent)
		readType()
		switch asset.AssetType {
		case 2:
			texture := MushTextureAsset{}
			reader.Seek(34, io.SeekCurrent)
			readType()
			reader.Seek(20, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &texture.Width)
			reader.Seek(18, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &texture.Height)
			reader.Seek(14, io.SeekCurrent)
			readType()
			reader.Seek(5, io.SeekCurrent)
			texture.Data = make([]byte, util.Read[int32](reader))
			reader.Seek(4, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &texture.Data)
			reader.Seek(27, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &texture.FilterMode)
			reader.Seek(22, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &texture.MipChain)
			reader.Seek(12, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &texture.HDR)
			reader.Seek(18, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &texture.Linear)
			reader.Seek(4, io.SeekCurrent)
			asset.AssetData = texture
		case 3:
			mesh := MushMeshAsset{}
			reader.Seek(28, io.SeekCurrent)
			readType()
			reader.Seek(26, io.SeekCurrent)
			readType()
			reader.Seek(5, io.SeekCurrent)
			mesh.Vertices = make([][3]float32, util.Read[uint64](reader))
			reader.Seek(1, io.SeekCurrent)
			readType()
			for i := range mesh.Vertices {
				reader.Seek(1, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &mesh.Vertices[i][0])
				reader.Seek(1, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &mesh.Vertices[i][1])
				reader.Seek(1, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &mesh.Vertices[i][2])
				reader.Seek(3, io.SeekCurrent)
				if i != len(mesh.Vertices)-1 {
					reader.Seek(4, io.SeekCurrent)
				}
			}
			reader.Seek(42, io.SeekCurrent)
			readType()
			reader.Seek(5, io.SeekCurrent)
			mesh.SubMeshDescriptors = make([]MushMeshDescriptor, util.Read[uint64](reader))
			reader.Seek(1, io.SeekCurrent)
			readType()
			for i := range mesh.SubMeshDescriptors {
				reader.Seek(30, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &mesh.SubMeshDescriptors[i].IndexStart)
				reader.Seek(26, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &mesh.SubMeshDescriptors[i].IndexCount)
				reader.Seek(38, io.SeekCurrent)
				readType()
				reader.Seek(5, io.SeekCurrent)
				theTriangles := &mesh.SubMeshDescriptors[i].SubMeshTriangles
				*theTriangles = make([]int32, util.Read[uint64](reader))
				for k := range *theTriangles {
					reader.Seek(1, io.SeekCurrent)
					binary.Read(reader, binary.LittleEndian, &(*theTriangles)[k])
				}
				reader.Seek(32, io.SeekCurrent)
				theIndices := &mesh.SubMeshDescriptors[i].Indices
				*theIndices = make([]int32, util.Read[uint64](reader))
				for k := range *theIndices {
					reader.Seek(1, io.SeekCurrent)
					binary.Read(reader, binary.LittleEndian, &(*theIndices)[k])
				}
				if i != len(mesh.SubMeshDescriptors)-1 {
					reader.Seek(9, io.SeekCurrent)
				}
			}
			reader.Seek(17, io.SeekCurrent)
			readType()
			reader.Seek(5, io.SeekCurrent)
			mesh.UVs = make([][2]float32, util.Read[uint64](reader))
			reader.Seek(1, io.SeekCurrent)
			readType()
			for i := range mesh.UVs {
				reader.Seek(1, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &mesh.UVs[i][0])
				reader.Seek(1, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &mesh.UVs[i][1])
				reader.Seek(3, io.SeekCurrent)
				if i != len(mesh.UVs)-1 {
					reader.Seek(4, io.SeekCurrent)
				}
			}
			reader.Seek(30, io.SeekCurrent)
			mesh.Normals = make([][3]float32, util.Read[uint64](reader))
			reader.Seek(6, io.SeekCurrent)
			for i := range mesh.Normals {
				reader.Seek(1, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &mesh.Normals[i][0])
				reader.Seek(1, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &mesh.Normals[i][1])
				reader.Seek(1, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &mesh.Normals[i][2])
				reader.Seek(3, io.SeekCurrent)
				if i != len(mesh.Normals)-1 {
					reader.Seek(4, io.SeekCurrent)
				}
			}
			reader.Seek(22, io.SeekCurrent)
			readType()
			reader.Seek(5, io.SeekCurrent)
			mesh.Tangents = make([][4]float32, util.Read[uint64](reader))
			reader.Seek(1, io.SeekCurrent)
			readType()
			for i := range mesh.Tangents {
				reader.Seek(1, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &mesh.Tangents[i][0])
				reader.Seek(1, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &mesh.Tangents[i][1])
				reader.Seek(1, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &mesh.Tangents[i][2])
				reader.Seek(1, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &mesh.Tangents[i][3])
				reader.Seek(3, io.SeekCurrent)
				if i != len(mesh.Tangents)-1 {
					reader.Seek(4, io.SeekCurrent)
				}
			}
			reader.Seek(4, io.SeekCurrent)
			asset.AssetData = mesh
		case 4:
			material := MushMaterialAsset{}
			reader.Seek(44, io.SeekCurrent)
			readType()
			reader.Seek(34, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.MaterialType)
			reader.Seek(18, io.SeekCurrent)
			material.Albedo = util.ReadString(reader)
			reader.Seek(22, io.SeekCurrent)
			material.Emission = util.ReadString(reader)
			reader.Seek(18, io.SeekCurrent)
			material.Normal = util.ReadString(reader)
			reader.Seek(14, io.SeekCurrent)
			material.Mask = util.ReadString(reader)
			reader.Seek(28, io.SeekCurrent)
			readType()
			reader.Seek(1, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.AlbedoColor[0])
			reader.Seek(1, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.AlbedoColor[1])
			reader.Seek(1, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.AlbedoColor[2])
			reader.Seek(1, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.AlbedoColor[3])
			reader.Seek(33, io.SeekCurrent)
			readType()
			reader.Seek(1, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.EmissionColor[0])
			reader.Seek(1, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.EmissionColor[1])
			reader.Seek(1, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.EmissionColor[2])
			reader.Seek(1, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.EmissionColor[3])
			reader.Seek(41, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.EmissionIntensity)
			reader.Seek(46, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.EmissionIntensityMin)
			reader.Seek(46, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.EmissionTwinkleSpeed)
			reader.Seek(22, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.Metallic)
			reader.Seek(26, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.Smoothness)
			reader.Seek(10, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.AO)
			reader.Seek(28, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.NormalScale)
			reader.Seek(30, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.OverlayTrack)
			reader.Seek(40, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.OverlayNoiseScale)
			reader.Seek(40, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.OverlayNoiseBlend)
			reader.Seek(30, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.OverlayClamp)
			reader.Seek(40, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.OverlayProjection)
			reader.Seek(32, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.OverlayOffset)
			reader.Seek(30, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.WetnessClamp)
			reader.Seek(32, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.WetnessOffset)
			reader.Seek(60, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.WetnessSmoothnessMultiplier)
			reader.Seek(28, io.SeekCurrent)
			readType()
			reader.Seek(1, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.TilingScale[0])
			reader.Seek(1, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.TilingScale[1])
			reader.Seek(31, io.SeekCurrent)
			readType()
			reader.Seek(1, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.TilingOffset[0])
			reader.Seek(1, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.TilingOffset[1])
			reader.Seek(29, io.SeekCurrent)
			readType()
			reader.Seek(1, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.TilingSpeed[0])
			reader.Seek(1, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.TilingSpeed[1])
			reader.Seek(23, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.GlobalUV)
			reader.Seek(30, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.GlobalUVTile)
			reader.Seek(32, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.GlobalUVBlend)
			material.TopAlbedo = util.ReadNullOrStringEntry(reader)
			material.TopNormal = util.ReadNullOrStringEntry(reader)
			material.TopMask = util.ReadNullOrStringEntry(reader)
			reader.Seek(38, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.TransparencyType)
			reader.Seek(42, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.TransparencyDither)
			reader.Seek(42, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.AlphaClipThreshold)
			reader.Seek(30, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.AffectAlbedo)
			reader.Seek(30, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.AffectNormal)
			reader.Seek(28, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.AffectMetal)
			reader.Seek(22, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.AffectAO)
			reader.Seek(38, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.AffectSmoothness)
			reader.Seek(34, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.AffectEmission)
			reader.Seek(24, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &material.DrawOrder)
			reader.Seek(4, io.SeekCurrent)
			asset.AssetData = material
		case 5:
			reader.Seek(30, io.SeekCurrent)
			readType()
			reader.Seek(18, io.SeekCurrent)
			readType()
			reader.Seek(5, io.SeekCurrent)
			audioBytes := make([]byte, util.Read[int32](reader))
			reader.Seek(4, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &audioBytes)
			reader.Seek(5, io.SeekCurrent)
			asset.AssetData = audioBytes
		case 6:
			reader.Seek(18, io.SeekCurrent)
			asset.AssetData = util.ReadString(reader)
			reader.Seek(3, io.SeekCurrent)
		default:
			fmt.Println(`wtf`)
		}
		pack.Assets[key] = asset
		i++
	}
	return
}

func decodeBinaryXML(raw []byte) []xml.Token {
	reader := bytes.NewReader(raw)
	tokens, cache := []xml.Token{}, []xml.Name{}
	popCache := func() {
		tokens = append(tokens, xml.EndElement{Name: cache[len(cache)-1]})
		cache = slices.Delete(cache, len(cache)-1, len(cache))
	}

	for reader.Len() > 0 {
		typ, _ := reader.ReadByte()
		processType(typ, reader, popCache, &cache, &tokens)
	}
	return tokens
}

func processType(typ byte, reader *bytes.Reader, popCache func(), cache *[]xml.Name, tokens *[]xml.Token) {
	switch {
	case typ == 0x01:
		popCache()
	case typ == 0x03:
		pos, _ := reader.Seek(0, io.SeekCurrent)
		peek, _ := reader.ReadByte()
		for peek != 1 {
			peek, _ = reader.ReadByte()
		}
		recordType, _ := reader.ReadByte()
		recordLen, _ := reader.ReadByte()
		for range recordLen {
			cur, _ := reader.Seek(0, io.SeekCurrent)
			reader.Seek(pos, io.SeekStart)
			theTyp, _ := reader.ReadByte()
			processType(theTyp, reader, popCache, cache, tokens)
			reader.Seek(cur, io.SeekStart)
			processType(recordType, reader, popCache, cache, tokens)
		}
	case typ == 0x08:
		last := (*tokens)[len(*tokens)-1].(xml.StartElement)
		last.Attr = append(last.Attr, xml.Attr{Name: xml.Name{Local: `xmlns`}, Value: readAsciiString(reader)})
		(*tokens)[len(*tokens)-1] = last
	case typ == 0x09:
		last := (*tokens)[len(*tokens)-1].(xml.StartElement)
		last.Attr = append(last.Attr, xml.Attr{Name: xml.Name{Local: `xmlns:` + readAsciiString(reader)}, Value: readAsciiString(reader)})
		(*tokens)[len(*tokens)-1] = last
	case typ >= 0x26 && typ <= 0x3F:
		name := readAsciiString(reader)
		value, boo := readStringRecord(reader)
		if boo {
			fmt.Println(`怎么是true`)
		}
		last := (*tokens)[len(*tokens)-1].(xml.StartElement)
		last.Attr = append(last.Attr, xml.Attr{Name: xml.Name{Local: string(typ-0x26+'a') + `:` + name}, Value: value})
		(*tokens)[len(*tokens)-1] = last
	case typ == 0x40:
		name := xml.Name{Local: readAsciiString(reader)}
		*tokens = append(*tokens, xml.StartElement{Name: name})
		*cache = append(*cache, name)
	case typ >= 0x5E && typ <= 0x77:
		name := xml.Name{Local: string(typ-0x5E+'a') + `:` + readAsciiString(reader)}
		*tokens = append(*tokens, xml.StartElement{Name: name})
		*cache = append(*cache, name)
	case typ == 0x80 || typ == 0x81:
		*tokens = append(*tokens, xml.CharData([]byte{'0'}))
		if typ == 0x81 {
			popCache()
		}
	case typ == 0x82 || typ == 0x83:
		*tokens = append(*tokens, xml.CharData([]byte{'1'}))
		if typ == 0x83 {
			popCache()
		}
	case typ == 0x84 || typ == 0x85:
		*tokens = append(*tokens, xml.CharData([]byte{'f', 'a', 'l', 's', 'e'}))
		if typ == 0x85 {
			popCache()
		}
	case typ == 0x86 || typ == 0x87:
		*tokens = append(*tokens, xml.CharData([]byte{'t', 'r', 'u', 'e'}))
		if typ == 0x87 {
			popCache()
		}
	case typ == 0x88 || typ == 0x89:
		*tokens = append(*tokens, xml.CharData([]byte(strconv.FormatInt(int64(util.Read[int8](reader)), 10))))
		if typ == 0x89 {
			popCache()
		}
	case typ == 0x8A || typ == 0x8B:
		*tokens = append(*tokens, xml.CharData([]byte(strconv.FormatInt(int64(util.Read[int16](reader)), 10))))
		if typ == 0x8B {
			popCache()
		}
	case typ == 0x8C || typ == 0x8D:
		*tokens = append(*tokens, xml.CharData([]byte(strconv.FormatInt(int64(util.Read[int32](reader)), 10))))
		if typ == 0x8D {
			popCache()
		}
	case typ == 0x8E || typ == 0x8F:
		*tokens = append(*tokens, xml.CharData([]byte(strconv.FormatInt(util.Read[int64](reader), 10))))
		if typ == 0x8F {
			popCache()
		}
	case typ == 0x90 || typ == 0x91:
		*tokens = append(*tokens, xml.CharData([]byte(strconv.FormatFloat(float64(util.Read[float32](reader)), 'f', -1, 32))))
		if typ == 0x91 {
			popCache()
		}
	case typ == 0x92 || typ == 0x93:
		*tokens = append(*tokens, xml.CharData([]byte(strconv.FormatFloat(util.Read[float64](reader), 'f', -1, 64))))
		if typ == 0x93 {
			popCache()
		}
	case typ == 0x98 || typ == 0x99:
		leng, _ := reader.ReadByte()
		*tokens = append(*tokens, xml.CharData(util.ReadNBytes(reader, leng)))
		if typ == 0x99 {
			popCache()
		}
	case typ == 0x9A || typ == 0x9B:
		leng := util.Read[uint16](reader)
		*tokens = append(*tokens, xml.CharData(util.ReadNBytes(reader, leng)))
		if typ == 0x9B {
			popCache()
		}
	case typ == 0xB4 || typ == 0xB5:
		boolean, _ := reader.ReadByte()
		if boolean != 0 {
			*tokens = append(*tokens, xml.CharData([]byte{'t', 'r', 'u', 'e'}))
		} else {
			*tokens = append(*tokens, xml.CharData([]byte{'f', 'a', 'l', 's', 'e'}))
		}
		if typ == 0xB5 {
			popCache()
		}
	case typ == 0xB6 || typ == 0xB7:
		leng, _ := reader.ReadByte()
		u16s := make([]uint16, leng>>1)
		binary.Read(reader, binary.LittleEndian, &u16s)
		str := string(utf16.Decode(u16s))
		*tokens = append(*tokens, xml.CharData([]byte(str)))
		if typ == 0xB7 {
			popCache()
		}
	case typ == 0xB8 || typ == 0xB9:
		leng := util.Read[uint16](reader)
		u16s := make([]uint16, leng>>1)
		binary.Read(reader, binary.LittleEndian, &u16s)
		str := string(utf16.Decode(u16s))
		*tokens = append(*tokens, xml.CharData([]byte(str)))
		if typ == 0xB9 {
			popCache()
		}
	default:
		fmt.Println(`未知XML节点类型`, typ, `位于`, reader.Size()-int64(reader.Len()))
	}
}

func readAsciiString(r *bytes.Reader) string {
	len, _ := r.ReadByte()
	return string(util.ReadNBytes(r, len))
}

func readStringRecord(r *bytes.Reader) (string, bool) {
	typ, _ := r.ReadByte()
	switch typ {
	case 0x86, 0x87:
		return `true`, typ == 0x87
	case 0x92, 0x93:
		return strconv.FormatFloat(util.Read[float64](r), 'f', -1, 64), typ == 0x93
	case 0x98, 0x99:
		return readAsciiString(r), typ == 0x99
	default:
		fmt.Println(`未知string record类型`, typ, `位于`, r.Size()-int64(r.Len()))
	}
	return ``, false
}
