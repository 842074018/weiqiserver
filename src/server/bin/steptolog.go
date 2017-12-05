package bin

import (
	"log"
)

func StepToGameInfo(gameStep []int64) [][]uint32 {
	gameSize := len(gameStep)
	gameInfo := make([][]uint32, gameSize)
	for k := range gameInfo {
		gameInfo[k] = make([]uint32, gameSize)
	}
	for i := 0; i < gameSize; i++ {
		row := gameStep[i]
		for j := uint(0); j < uint(gameSize*2); j += 2 {
			var (
				b int64
				c int64
			)
			var (
				e uint64
				f uint64
			)
			if j == 0 {
				d := uint64(row)
				e = d << 62
				f = e >> 62
				c = int64(f)
			} else {
				b = row << (61 - j)
				c = b >> (61)
			}
			switch c {
			case 0:
				//空白子
				gameInfo[i][j/2] = 0
			case 2:
				//黑子
				gameInfo[i][j/2] = 1
			case 3:
				//白子
				gameInfo[i][j/2] = 2
			}
		}
	}
	return gameInfo
}

func StepLogToGameShow(gameStep [][]uint32) []int64 {
	sizeLen := len(gameStep)
	newJoinLog := make([]int64, sizeLen)
	for i := 0; i < sizeLen; i++ {
		newLog := int64(0)
		log.Println("gameStep", gameStep[i])
		for j := uint(0); j < uint(sizeLen*2); j += 2 {
			switch gameStep[i][j/2] {
			// case 0:
			// 	newLog |= 0 << j * 2
			// 	newLog |= 0<<j*2 + 1
			case 1:
				newLog |= 1 << (j + 1)
				newLog |= 0 << j
			case 2:
				newLog |= 1 << j
				newLog |= 1 << (j + 1)
			}
		}
		log.Println("exchange data", newLog)
		newJoinLog[i] = newLog
	}
	return newJoinLog
}
