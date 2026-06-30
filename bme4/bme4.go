package bme4

import (
	"BallexSeriesCustomMapDataAnalyzer/ballex2"
	"BallexSeriesCustomMapDataAnalyzer/util"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type ItemType uint64

const (
	PhysicsObject ItemType = iota
	Folder
	Light
	TransformMark
	SkyBox
	Joint
	RoadGenerator
	ImageLoader
	ItemLink
	Waypoint
	WayPath
	Trigger
	Executor
	Variable
	ParticleSystem
	VegetationGenerator
	TerrainGenerator
	Terrain
	AudioPlayer
	CollectableRegister
	CollectableObject
	CustomExecutor
	Fog
	AssetReference
	Camera
	UI
)

type BMS struct {
	BMSInfo   BMSInfo
	BMSItems  []BMSItem
	BMSAssets []BMSAsset
}

type BMSInfo struct {
	EditorVersion     int32
	MapId             string
	AuthorName        string
	AuthorSteamId     uint64
	LevelName         string
	LevelVersionMajor int32
	LevelVersionMinor int32
	LevelVersionPatch int32
	LevelDifficulty   int32
	LevelDescription  string
	InitialBallType   int32
	Cover             *MushTextureAsset
	CameraMode        int32
	CameraOffset      [3]float32
	Gravity           [3]float32
	NoGravity         bool
	EnvironmentTemp   float32
	ViewDistance      float32
}

type MushTextureAsset struct {
	Width, Height int32
	Data          []byte
}

type BMSItem struct {
	Name     string
	Id       int32
	Trans    Trans
	ItemType ItemType
	ItemData ItemData
	Template bool
}

type Trans struct {
	Pos, Rot, Scale [3]float32
}

type ItemData struct {
	DataVersion           int32
	IntDictionary         map[string]int32
	IntArrayDictionary    map[string][]int32
	FloatDictionary       map[string]float32
	FloatArrayDictionary  map[string][]float32
	DoubleArrayDictionary map[string][]float64
	BoolDictionary        map[string]bool
	BoolArrayDictionary   map[string][]bool
	StringDictionary      map[string]string
	StringArrayDictionary map[string][]string
	VectorDictionary      map[string][4]float32
	VectorArrayDictionary map[string][][4]float32
}

type BMSAsset struct {
	Name             string
	AssetType        uint64
	IsBuiltInAsset   bool
	BuiltInAssetLink string
	AssetData        any
}

type MushMaterialAsset struct {
	Albedo, Normal, Mask                   string
	GlobalUV                               bool
	TilingScale, TilingOffset, TilingSpeed [2]float32
	EmissionColor                          [4]float32
	Metallic, Smoothness, AO               float32
	TransparencyType                       int32
	BlendMode                              int32
	MaterialType                           int32
}

func ReadBME4MapData(raw []byte) {
	bms := BMS{}
	reader := bytes.NewReader(raw)
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

	reader.Seek(217, io.SeekStart)
	binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.EditorVersion)
	reader.Seek(16, io.SeekCurrent)
	bms.BMSInfo.MapId = util.ReadString(reader)
	reader.Seek(26, io.SeekCurrent)
	bms.BMSInfo.AuthorName = util.ReadString(reader)
	reader.Seek(32, io.SeekCurrent)
	binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.AuthorSteamId)
	reader.Seek(24, io.SeekCurrent)
	bms.BMSInfo.LevelName = util.ReadString(reader)
	reader.Seek(40, io.SeekCurrent)
	binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.LevelVersionMajor)
	reader.Seek(40, io.SeekCurrent)
	binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.LevelVersionMinor)
	reader.Seek(40, io.SeekCurrent)
	binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.LevelVersionPatch)
	reader.Seek(36, io.SeekCurrent)
	binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.LevelDifficulty)
	reader.Seek(38, io.SeekCurrent)
	bms.BMSInfo.LevelDescription = util.ReadString(reader)
	reader.Seek(36, io.SeekCurrent)
	binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.InitialBallType)
	reader.Seek(16, io.SeekCurrent)
	if tmp, _ := reader.ReadByte(); tmp == 47 {
		bms.BMSInfo.Cover = new(MushTextureAsset)
		reader.Seek(-1, io.SeekCurrent)
		readType()
		reader.Seek(20, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.Cover.Width)
		reader.Seek(18, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.Cover.Height)
		reader.Seek(14, io.SeekCurrent)
		readType()
		reader.Seek(5, io.SeekCurrent)
		bms.BMSInfo.Cover.Data = make([]byte, util.Read[int32](reader))
		reader.Seek(4, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.Cover.Data)
		reader.Seek(2, io.SeekCurrent)
	}
	reader.Seek(26, io.SeekCurrent)
	binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.CameraMode)
	reader.Seek(30, io.SeekCurrent)
	readType()
	reader.Seek(1, io.SeekCurrent)
	binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.CameraOffset[0])
	reader.Seek(1, io.SeekCurrent)
	binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.CameraOffset[1])
	reader.Seek(1, io.SeekCurrent)
	binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.CameraOffset[2])
	reader.Seek(27, io.SeekCurrent)
	binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.Gravity[0])
	reader.Seek(1, io.SeekCurrent)
	binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.Gravity[1])
	reader.Seek(1, io.SeekCurrent)
	binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.Gravity[2])
	reader.Seek(25, io.SeekCurrent)
	binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.NoGravity)
	if tmp, _ := reader.ReadByte(); tmp != 5 {
		reader.Seek(35, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.EnvironmentTemp)
		if tmp, _ := reader.ReadByte(); tmp != 5 {
			reader.Seek(29, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &bms.BMSInfo.ViewDistance)
			reader.Seek(1, io.SeekCurrent)
		}
	}

	reader.Seek(203, io.SeekCurrent)
	bms.BMSItems = make([]BMSItem, util.Read[uint64](reader))
	firstComparer := true
	for i := range bms.BMSItems {
		fmt.Println(`物体`, i, `位于`, reader.Size()-int64(reader.Len()))
		item := &bms.BMSItems[i]
		reader.Seek(1, io.SeekCurrent)
		readType()
		reader.Seek(18, io.SeekCurrent)
		item.Name = util.ReadString(reader)
		reader.Seek(10, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &item.Id)
		reader.Seek(16, io.SeekCurrent)
		readType()
		reader.Seek(18, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &item.Trans.Pos[0])
		reader.Seek(1, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &item.Trans.Pos[1])
		reader.Seek(1, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &item.Trans.Pos[2])
		reader.Seek(19, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &item.Trans.Rot[0])
		reader.Seek(1, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &item.Trans.Rot[1])
		reader.Seek(1, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &item.Trans.Rot[2])
		reader.Seek(19, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &item.Trans.Scale[0])
		reader.Seek(1, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &item.Trans.Scale[1])
		reader.Seek(1, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &item.Trans.Scale[2])
		reader.Seek(24, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &item.ItemType)
		reader.Seek(22, io.SeekCurrent)
		readType()
		reader.Seek(32, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &item.ItemData.DataVersion)
		reader.Seek(1, io.SeekCurrent)
		for {
			propName := util.ReadString(reader)
			readType()
			reader.Seek(26, io.SeekCurrent)
			if firstComparer {
				readType()
				reader.Seek(1, io.SeekCurrent)
			}
			reader.Seek(5, io.SeekCurrent)
			count := util.Read[uint64](reader)
			switch propName {
			case `intDictionary`:
				item.ItemData.IntDictionary = make(map[string]int32, count)
			case `intArrayDictionary`:
				item.ItemData.IntArrayDictionary = make(map[string][]int32, count)
			case `floatDictionary`:
				item.ItemData.FloatDictionary = make(map[string]float32, count)
			case `floatArrayDictionary`:
				item.ItemData.FloatArrayDictionary = make(map[string][]float32, count)
			case `doubleArrayDictionary`:
				item.ItemData.DoubleArrayDictionary = make(map[string][]float64, count)
			case `boolDictionary`:
				item.ItemData.BoolDictionary = make(map[string]bool, count)
			case `boolArrayDictionary`:
				item.ItemData.BoolArrayDictionary = make(map[string][]bool, count)
			case `stringDictionary`:
				item.ItemData.StringDictionary = make(map[string]string, count)
			case `stringArrayDictionary`:
				item.ItemData.StringArrayDictionary = make(map[string][]string, count)
			case `vectorDictionary`:
				item.ItemData.VectorDictionary = make(map[string][4]float32, count)
			case `vectorArrayDictionary`:
				item.ItemData.VectorArrayDictionary = make(map[string][][4]float32, count)
			}
			for range count {
				reader.Seek(12, io.SeekCurrent)
				key := util.ReadString(reader)
				reader.Seek(10, io.SeekCurrent)
				switch propName {
				case `intDictionary`:
					item.ItemData.IntDictionary[key] = util.Read[int32](reader)
					reader.Seek(1, io.SeekCurrent)
				case `intArrayDictionary`:
					readType()
					reader.Seek(5, io.SeekCurrent)
					tmp := make([]int32, util.Read[int32](reader))
					reader.Seek(4, io.SeekCurrent)
					binary.Read(reader, binary.LittleEndian, &tmp)
					item.ItemData.IntArrayDictionary[key] = tmp
					reader.Seek(2, io.SeekCurrent)
				case `floatDictionary`:
					item.ItemData.FloatDictionary[key] = util.Read[float32](reader)
					reader.Seek(1, io.SeekCurrent)
				case `floatArrayDictionary`:
					readType()
					reader.Seek(5, io.SeekCurrent)
					tmp := make([]float32, util.Read[int32](reader))
					reader.Seek(4, io.SeekCurrent)
					binary.Read(reader, binary.LittleEndian, &tmp)
					item.ItemData.FloatArrayDictionary[key] = tmp
					reader.Seek(2, io.SeekCurrent)
				case `doubleArrayDictionary`:
					readType()
					reader.Seek(5, io.SeekCurrent)
					tmp := make([]float64, util.Read[int32](reader))
					reader.Seek(4, io.SeekCurrent)
					binary.Read(reader, binary.LittleEndian, &tmp)
					item.ItemData.DoubleArrayDictionary[key] = tmp
					reader.Seek(2, io.SeekCurrent)
				case `boolDictionary`:
					item.ItemData.BoolDictionary[key] = util.Read[bool](reader)
					reader.Seek(1, io.SeekCurrent)
				case `boolArrayDictionary`:
					readType()
					reader.Seek(5, io.SeekCurrent)
					tmp := make([]bool, util.Read[int32](reader))
					reader.Seek(4, io.SeekCurrent)
					binary.Read(reader, binary.LittleEndian, &tmp)
					item.ItemData.BoolArrayDictionary[key] = tmp
					reader.Seek(2, io.SeekCurrent)
				case `stringDictionary`:
					item.ItemData.StringDictionary[key] = util.ReadString(reader)
					reader.Seek(1, io.SeekCurrent)
				case `stringArrayDictionary`:
					readType()
					reader.Seek(5, io.SeekCurrent)
					tmp := make([]string, util.Read[uint64](reader))
					for i := range tmp {
						reader.Seek(1, io.SeekCurrent)
						tmp[i] = util.ReadString(reader)
					}
					item.ItemData.StringArrayDictionary[key] = tmp
					reader.Seek(3, io.SeekCurrent)
				case `vectorDictionary`:
					readType()
					tmp := [4]float32{}
					reader.Seek(1, io.SeekCurrent)
					binary.Read(reader, binary.LittleEndian, &tmp[0])
					reader.Seek(1, io.SeekCurrent)
					binary.Read(reader, binary.LittleEndian, &tmp[1])
					reader.Seek(1, io.SeekCurrent)
					binary.Read(reader, binary.LittleEndian, &tmp[2])
					reader.Seek(1, io.SeekCurrent)
					binary.Read(reader, binary.LittleEndian, &tmp[3])
					item.ItemData.VectorDictionary[key] = tmp
					reader.Seek(2, io.SeekCurrent)
				case `vectorArrayDictionary`:
					readType()
					reader.Seek(5, io.SeekCurrent)
					tmp := make([][4]float32, util.Read[uint64](reader))
					for i := range tmp {
						reader.Seek(1, io.SeekCurrent)
						readType()
						reader.Seek(1, io.SeekCurrent)
						binary.Read(reader, binary.LittleEndian, &tmp[i][0])
						reader.Seek(1, io.SeekCurrent)
						binary.Read(reader, binary.LittleEndian, &tmp[i][1])
						reader.Seek(1, io.SeekCurrent)
						binary.Read(reader, binary.LittleEndian, &tmp[i][2])
						reader.Seek(1, io.SeekCurrent)
						binary.Read(reader, binary.LittleEndian, &tmp[i][3])
						reader.Seek(1, io.SeekCurrent)
					}
					item.ItemData.VectorArrayDictionary[key] = tmp
					reader.Seek(3, io.SeekCurrent)
				}
			}
			if util.ReadNBytes(reader, 3)[2] != 1 {
				break
			}
			firstComparer = false
		}
		if tmp, _ := reader.ReadByte(); tmp != 5 {
			reader.Seek(21, io.SeekCurrent)
			binary.Read(reader, binary.LittleEndian, &item.Template)
			reader.Seek(1, io.SeekCurrent)
		}
	}
	reader.Seek(209, io.SeekCurrent)
	bms.BMSAssets = make([]BMSAsset, util.Read[uint64](reader))
	for i := range bms.BMSAssets {
		fmt.Println(`资源`, i, `位于`, reader.Size()-int64(reader.Len()))
		asset := &bms.BMSAssets[i]
		reader.Seek(1, io.SeekCurrent)
		readType()
		reader.Seek(18, io.SeekCurrent)
		asset.Name = util.ReadString(reader)
		reader.Seek(24, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &asset.AssetType)
		tmp, _ := reader.ReadByte()
		reader.Seek(23, io.SeekCurrent)
		if tmp == 1 {
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
				reader.Seek(3, io.SeekCurrent)
				asset.AssetData = texture
			case 3:
				mesh := ballex2.MushMeshAsset{}
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
				mesh.SubMeshDescriptors = make([]ballex2.MushMeshDescriptor, util.Read[uint64](reader))
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
				reader.Seek(2, io.SeekCurrent)
				asset.AssetData = mesh
			case 4:
				material := MushMaterialAsset{}
				reader.Seek(44, io.SeekCurrent)
				readType()
				reader.Seek(4, io.SeekCurrent)
				material.Albedo = util.ReadNullOrStringEntry(reader)
				material.Normal = util.ReadNullOrStringEntry(reader)
				material.Mask = util.ReadNullOrStringEntry(reader)
				reader.Seek(22, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &material.GlobalUV)
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
				reader.Seek(23, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &material.Metallic)
				reader.Seek(26, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &material.Smoothness)
				reader.Seek(10, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &material.AO)
				reader.Seek(38, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &material.TransparencyType)
				reader.Seek(24, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &material.BlendMode)
				reader.Seek(30, io.SeekCurrent)
				binary.Read(reader, binary.LittleEndian, &material.MaterialType)
				reader.Seek(2, io.SeekCurrent)
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
				reader.Seek(3, io.SeekCurrent)
				asset.AssetData = audioBytes
			}
		}
		if i == 17 {
			fmt.Println()
		}
		reader.Seek(34, io.SeekCurrent)
		binary.Read(reader, binary.LittleEndian, &asset.IsBuiltInAsset)
		asset.BuiltInAssetLink = util.ReadNullOrStringEntry(reader)
		reader.Seek(1, io.SeekCurrent)
		// pos := reader.Size() - int64(reader.Len())
		// fmt.Println(pos)
	}
	fmt.Println(`我去！程序顺利结束了，没有崩！打断点看结构体信息吧`)
}
