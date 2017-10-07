package main

import (
	"fmt"
	"sort"
)

type SortBets []int

func (s SortBets) Len() int {
	return len(s)
}
func (s SortBets) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s SortBets) Less(i, j int) bool {
	return s[i] < s[j]
}

type Player struct {
	Balance int   `json:"balance"`
	Bets    []int `json:"bet"`
}

type PlasersState struct {
	Name      string `json:"name"`
	Balance   int    `json:"balance"`
	NumOfBets int    `json:"numofbets"`
}

func (p *Player) resetBets() int {
	totalBets := 0
	for _, v := range p.Bets {
		if v < 1 {
			continue
		}
		totalBets = totalBets + v
	}
	p.Bets = []int{}
	p.Balance = p.Balance - totalBets
	return totalBets
}

func (p *Player) WinnerAddBalance(bets int) {
	p.Balance = p.Balance + bets
}

func FindWinner(board Board) (bool, Board) {
	fmt.Println("find winner pirnt  board")
	fmt.Println(board)
	allBetsArr := SortBets{}
	for _, v := range board.Players {
		allBetsArr = append(allBetsArr, v.Bets...)
	}
	sort.Sort(allBetsArr)
	fmt.Println("sort")
	fmt.Println(allBetsArr)
	onceBetsArr := removeDuplicatesAndEmpty(allBetsArr)
	if len(onceBetsArr) >= 1 {
		return true, winPotBySmallestNum(board, onceBetsArr[0])
	}
	return false, noWinnerGetBackBet(board)
}

func winPotBySmallestNum(b Board, num int) Board {
	fmt.Println("win pot  print board")
	fmt.Println(b, num)

	winName := ""
	for name, v := range b.Players {
		for _, bet := range v.Bets {
			if num == bet {
				winName = name
			}
		}
	}

	totalBets := 0
	for name, v := range b.Players {
		totalBets = totalBets + v.resetBets()
		b.putPlayerByName(name, v)
	}
	winner := b.getPlayerByName(winName)
	winner.WinnerAddBalance(totalBets)
	b.putPlayerByName(winName, winner)
	return b
}

func noWinnerGetBackBet(b Board) Board {
	for name, v := range b.Players {
		v.Bets = []int{}
		b.putPlayerByName(name, v)
	}
	return b
}

func removeDuplicatesAndEmpty(a SortBets) []int {
	l := len(a)
	if l <= 1 {
		return a
	}

	dupArr := []int{}
	for i := 0; i < l-1; i++ {
		if a[i] == a[i+1] {
			dupArr = append(dupArr, a[i])
		}
	}
	//Delete duplicate value
	for i := 0; i < len(a); i++ {
		for _, dup := range dupArr {
			if a[i] == dup || a[i] < 1 {
				a = append(a[:i], a[i+1:]...)
				i--
				break
			}
		}
	}

	fmt.Println("remove duplicate and empty")
	fmt.Println(a)
	return a
}
