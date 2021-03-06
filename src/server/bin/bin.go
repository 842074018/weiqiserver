package bin

import (
	"center"
	"conf"
	"fmt"
	"log"
	"server/cache"
	"server/db"
	"server/game"
	"strconv"
)

func Weiqi01(playerId string) *game.RESP_Weiqi_01 {
	player, err := db.GetPlayerInfo(playerId)
	if err != nil {
		log.Println("The first time to login in:", playerId)
		player = &game.PlayerInfo{}
		player.Default(playerId)
	}
	// add into PlayerList
	err = db.SetAllPlayerIdList(playerId)
	if err != nil {
		log.Println(err)
		return &game.RESP_Weiqi_01{
			Status: conf.ERR_SERVER_ERR,
		}
	}
	// save playerinfo
	err = db.SetPlayerInfo(player.GetDbKey(), player)
	if err != nil {
		log.Println(err)
		return &game.RESP_Weiqi_01{
			Status: conf.ERR_SERVER_ERR,
		}
	}
	cache.OutAddOnlinePlayer(player.PlayerId)
	onlineList := cache.GetAllOnlinePlayer(playerId)
	return &game.RESP_Weiqi_01{
		Status:       conf.SUCCEED,
		OnlinePlayer: onlineList,
	}
}

func Weiqi02(playerId string) *game.RESP_Weiqi_02 {
	player, err := db.GetPlayerInfo(playerId)
	if err != nil {
		log.Println("Never login", playerId)
		return &game.RESP_Weiqi_02{
			Status: conf.ERR_SERVER_ERR,
		}
	}
	log.Println("PlayerInfo:", player)
	// keep alive
	err = db.SetAllPlayerIdList(playerId)
	if err != nil {
		log.Println(err)
		return &game.RESP_Weiqi_02{
			Status: conf.ERR_SERVER_ERR,
		}
	}
	// get all onlineplayer
	cache.OutAddOnlinePlayer(player.PlayerId)
	onlineList := cache.GetAllOnlinePlayer(playerId)
	liveGame := player.GetOnGame()
	// add gameInfo
	allGameInfo := GetAllOnlineGameInfo(liveGame)
	isEnd := cache.GetMatchStatusByPlayerId(playerId)
	return &game.RESP_Weiqi_02{
		Status:       conf.SUCCEED,
		OnlinePlayer: onlineList,
		AllGameInfo:  allGameInfo,
		InviteInfo:   isEnd,
	}
}

func Weiqi03(playerId string, inviteId string, size int) *game.RESP_Weiqi_03 {
	log.Println("playerId:", playerId, "iId:", inviteId, "size:", size)
	player, err := db.GetPlayerInfo(playerId)
	if err != nil {
		log.Println("Never login", playerId)
		return &game.RESP_Weiqi_03{
			Status: conf.ERR_SERVER_ERR,
		}
	}
	// TODO check is online
	// check inviteid is alive
	// if !cache.IsPlayerOnline(inviteId) {
	// 	log.Println("Invite player is offline:", inviteId)
	// 	return &game.RESP_Weiqi_03{
	// 		Status: conf.ERR_INVITE_OFFLINE,
	// 	}
	// }
	invitePlayer, err := db.GetPlayerInfo(inviteId)
	if err != nil {
		log.Println("Never login", inviteId)
		return &game.RESP_Weiqi_03{
			Status: conf.ERR_SERVER_ERR,
		}
	}
	// check size
	if size != conf.WEIQI_SIZE_SMALL && size != conf.WEIQI_SIZE_MID && size != conf.WEIQI_SIZE_STANDARD {
		return &game.RESP_Weiqi_03{
			Status: conf.ERR_BAD_PARAM,
		}
	}
	// make a new game
	playerList := []string{playerId, inviteId}
	gameInfo := game.NewOneGame(playerList, size)
	// save to db
	err = db.SetRedisC(gameInfo.GetDbKey(), gameInfo)
	if err != nil {
		log.Println("Set failed", err)
		return &game.RESP_Weiqi_03{
			Status: conf.ERR_SERVER_ERR,
		}
	}
	// add gameinfo
	player.JoinNewGameWithColor(gameInfo)
	invitePlayer.JoinNewGameWithColor(gameInfo)
	// save to db
	err = db.SetPlayerInfo(player.GetDbKey(), player)
	if err != nil {
		log.Println(err)
		return &game.RESP_Weiqi_03{
			Status: conf.ERR_SERVER_ERR,
		}
	}
	err = db.SetPlayerInfo(invitePlayer.GetDbKey(), invitePlayer)
	if err != nil {
		log.Println(err)
		return &game.RESP_Weiqi_03{
			Status: conf.ERR_SERVER_ERR,
		}
	}
	return &game.RESP_Weiqi_03{
		Status: conf.SUCCEED,
		GameId: gameInfo.WeiqiId,
	}
}

func Weiqi04(playerId string, gameId string, nextStep int) *game.RESP_Weiqi_04 {
	_, err := db.GetPlayerInfo(playerId)
	if err != nil {
		log.Println("Bad PlayerId", playerId)
		return &game.RESP_Weiqi_04{
			Status: conf.ERR_SERVER_ERR,
		}
	}
	key := fmt.Sprintf("Weiqi:Game:%v", gameId)
	gameInfo, err := db.GetRedisC(key)
	if err != nil {
		log.Println("Bad GameId", gameId)
		return &game.RESP_Weiqi_04{
			Status: conf.ERR_SERVER_ERR,
		}
	}
	log.Println("gameInfo:", gameInfo)
	//check is the right step
	nextStepColor := gameInfo.GetNextStepColor()
	playerColor, _ := gameInfo.GetWeiqiPlayerColor(playerId)
	if nextStepColor != playerColor {
		log.Println("Bad Game Step")
		return &game.RESP_Weiqi_04{
			Status: conf.ERR_BAD_PARAM,
		}
	}
	gameInfo.AddOneLogStep(nextStepColor, nextStep)
	y := nextStep / 19
	x := nextStep % 19
	log.Print("nextstep:", nextStep, "xy:", x, y)
	// JoinLog change to [size][size]uint32
	gameLogStep := StepToGameInfo(gameInfo.JoinLog)
	if nextStep != conf.GIVE_UP {
		gameLogStep[x][y] = nextStepColor + 1
	}
	// 进行提子
	log.Println("oldGame:", gameLogStep)
	newGameLogStep := center.GameCenterLogic(gameLogStep, nextStepColor, gameInfo.Size)
	//log.Println("newGame:", newGameLogStep)
	//newGameLogStep := gameLogStep
	newJoinStep := StepLogToGameShow(newGameLogStep)
	gameInfo.JoinLog = newJoinStep
	//save db
	// save to db
	err = db.SetRedisC(gameInfo.GetDbKey(), gameInfo)
	if err != nil {
		log.Println("Set failed", err)
		return &game.RESP_Weiqi_04{
			Status: conf.ERR_SERVER_ERR,
		}
	}
	return &game.RESP_Weiqi_04{
		Status:     conf.SUCCEED,
		GameStatus: newJoinStep,
	}
}

func Weiqi05(playerId string, gameId string, typeId string) uint32 {
	playerInfo, err := db.GetPlayerInfo(playerId)
	if err != nil {
		log.Println("Bad PlayerId", playerId)
		return conf.ERR_SERVER_ERR
	}
	key := fmt.Sprintf("Weiqi:Game:%v", gameId)
	gameInfo, err := db.GetRedisC(key)
	if err != nil {
		log.Println("Bad GameId", gameId)
		return conf.ERR_SERVER_ERR
	}
	color, _ := gameInfo.GetWeiqiPlayerColor(playerId)
	winColor := uint32(0)
	switch color {
	case conf.BLACK_PLAYER:
		winColor = conf.WHITE_PLAYER
	case conf.WHITE_PLAYER:
		winColor = conf.BLACK_PLAYER
	}
	gameInfo.IsEnd = true
	gameInfo.Winner = winColor + 1
	gameIdNum := gameInfo.WeiqiId
	playerInfo.AllWQ[gameIdNum] = []uint32{color, 2}
	//set player winner
	for _, v := range gameInfo.Player {
		if v != playerId {
			//is winner player
			winnerPlayer, err := db.GetPlayerInfo(v)
			if err != nil {
				log.Println("Bad PlayerId", v)
				return conf.ERR_SERVER_ERR
			}
			winnerPlayer.AllWQ[gameIdNum] = []uint32{winColor, 1}
			// SAVE DB
			err = db.SetPlayerInfo(winnerPlayer.GetDbKey(), winnerPlayer)
			if err != nil {
				log.Println(err)
				return conf.ERR_SERVER_ERR
			}
		}
	}
	err = db.SetPlayerInfo(playerInfo.GetDbKey(), playerInfo)
	if err != nil {
		log.Println(err)
		return conf.ERR_SERVER_ERR
	}
	err = db.SetRedisC(gameInfo.GetDbKey(), gameInfo)
	if err != nil {
		log.Println("Set failed", err)
		return conf.ERR_SERVER_ERR
	}
	return conf.SUCCEED
}

func Weiqi06(playId string, gameId string) *game.RESP_Weiqi_06 {
	_, err := db.GetPlayerInfo(playId)
	if err != nil {
		log.Println("Bad PlayerId", playId)
		return &game.RESP_Weiqi_06{
			Status: conf.ERR_SERVER_ERR,
		}
	}
	key := fmt.Sprintf("Weiqi:Game:%v", gameId)
	gameInfo, err := db.GetRedisC(key)
	if err != nil {
		log.Println("Bad GameId", gameId)
		return &game.RESP_Weiqi_06{
			Status: conf.ERR_SERVER_ERR,
		}
	}
	roundColor := gameInfo.GetNextStepColor()
	playInfo := gameInfo.Player
	size := gameInfo.Size
	gameStatus := gameInfo.JoinLog
	return &game.RESP_Weiqi_06{
		Status:     conf.SUCCEED,
		Round:      roundColor,
		Player:     playInfo,
		Size:       size,
		GameStatus: gameStatus,
	}
}

func Weiqi07(playerId string, matchType string, gameSize string) uint32 {
	_, err := db.GetPlayerInfo(playerId)
	if err != nil {
		log.Println("Bad PlayerId", playerId)
		return conf.ERR_SERVER_ERR
	}
	sizeNum, _ := strconv.Atoi(gameSize)
	if matchType == "0" {
		cache.AddOnePlayerBySize(playerId, sizeNum)
		return conf.SUCCEED
	} else if matchType == "1" {
		cache.EndMatchBySize(playerId, sizeNum)
		return conf.SUCCEED
	}
	return conf.ERR_BAD_PARAM
}
