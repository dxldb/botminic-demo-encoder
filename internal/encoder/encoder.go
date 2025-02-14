package encoder

import (
	"bytes"
	"os"
	"strconv"
	"time"

	ilog "botminic-demo-encoder/internal/logger"
)

const __MAGIC__ int32 = -559038737
const __FORMAT_VERSION__ int8 = 2

var bufMap = make(map[uint64]*bytes.Buffer)
var PlayerFramesMap = make(map[uint64][]FrameInfo)

var saveDir string = "./output"

func init() {
	if ok, _ := PathExists(saveDir); !ok {
		os.Mkdir(saveDir, os.ModePerm)
		ilog.InfoLogger.Println("未找到保存目录，已创建：", saveDir)
	} else {
		ilog.InfoLogger.Println("保存目录存在：", saveDir)
	}
}

func InitPlayer(initFrame FrameInitInfo, realTick int) {
	if bufMap[initFrame.PlayerSteamId64] == nil {
		bufMap[initFrame.PlayerSteamId64] = new(bytes.Buffer)
	} else {
		bufMap[initFrame.PlayerSteamId64].Reset()
	}
	// step.1 MAGIC NUMBER
	WriteToBuf(initFrame.PlayerSteamId64, __MAGIC__)

	// step.2 VERSION
	WriteToBuf(initFrame.PlayerSteamId64, __FORMAT_VERSION__)

	// step.3 demo tickrate
	WriteToBuf(initFrame.PlayerSteamId64, int16(realTick))

	// step.3 timestamp
	WriteToBuf(initFrame.PlayerSteamId64, int32(time.Now().Unix()))

	// step.4 name length
	WriteToBuf(initFrame.PlayerSteamId64, uint8(len(initFrame.PlayerName)))

	// step.5 name
	WriteToBuf(initFrame.PlayerSteamId64, []byte(initFrame.PlayerName))

	ilog.InfoLogger.Println("初始化成功: ", initFrame.PlayerName)
}

func WriteToRecFile(playerName string, playerSteamId64 uint64, roundNum int32, team string, uniqueID int32) {
	//if roundNum > 3 {
	//	return
	//}

	subDir := saveDir + "/round" + strconv.Itoa(int(roundNum)) + "/"
	if ok, _ := PathExists(subDir); !ok {
		os.MkdirAll(subDir, os.ModePerm)
		ilog.InfoLogger.Println(subDir)
	}

	fileName := subDir + strconv.FormatUint(uint64(uniqueID), 10) + "_" + team + ".rec"
	file, err := os.Create(fileName) // 创建文件, "binbin"是文件名字
	if err != nil {
		ilog.ErrorLogger.Println("文件创建失败", err.Error())
		return
	} else {
		ilog.InfoLogger.Println(fileName)
	}

	defer file.Close()

	// step.6 tick count
	var tickCount = int32(len(PlayerFramesMap[playerSteamId64])) + 1

	WriteToBuf(playerSteamId64, tickCount)

	// step.7 all tick frame
	for _, frame := range PlayerFramesMap[playerSteamId64] {
		for idx := 0; idx < 3; idx++ {
			WriteToBuf(playerSteamId64, frame.Origin[idx])
		}

		for idx := 0; idx < 2; idx++ {
			WriteToBuf(playerSteamId64, frame.Angle[idx])
		}

		WriteToBuf(playerSteamId64, frame.PlayerButtons)
		WriteToBuf(playerSteamId64, frame.EntityFlag)
		WriteToBuf(playerSteamId64, frame.MoveType)
		WriteToBuf(playerSteamId64, frame.CSWeaponID)
		// event_bomb_planted
		WriteToBuf(playerSteamId64, frame.Site)
		// event_item_drop
		WriteToBuf(playerSteamId64, frame.ItemDropped)
		// event_player_death
		WriteToBuf(playerSteamId64, frame.Victim)
		WriteToBuf(playerSteamId64, frame.Attacker)
		WriteToBuf(playerSteamId64, frame.HitGroup)
		// player props
		WriteToBuf(playerSteamId64, frame.Health)
		WriteToBuf(playerSteamId64, frame.Armor)
		WriteToBuf(playerSteamId64, frame.HasDefuser)
		WriteToBuf(playerSteamId64, frame.HasHelmet)
	}

	delete(PlayerFramesMap, playerSteamId64)
	file.Write(bufMap[playerSteamId64].Bytes())
	ilog.InfoLogger.Printf("[第%d回合] [%s] 录像保存成功: %d.rec\n", roundNum, playerName, playerSteamId64)
}
