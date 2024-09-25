package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/enescakir/emoji"
	"github.com/rivo/uniseg"
)

var SAVE_NAME = os.Getenv("HOME") + "/jackpot_save.json"

const DAILY_TOKENS = 40
const LOSE_JACKPOT_INCREASE = 5
const SPIN_TICKS = 90
const GAME_VERSION = "0.3"
const CHANCE_TO_WIN = 20
const TOKEN_COST = 100

type model struct {
	Tokens          int             `json:"tokens"`
	Dollars         int             `json:"dollars"`
	TicksLeft       int             `json:"ticksLeft"`
	SpinnerMsg      string          `json:"spinnerMsg"`
	Spinning        bool            `json:"spinning"`
	IsDay           bool            `json:"isDay"`
	JackpotAmount   int             `json:"jackpotAmount"`
	Slot            Slot            `json:"slot"`
	IsFreeSpin      bool            `json:"isFreeSpin"`
	FeverMode       bool            `json:"feverMode"`
	SymbolDropTable SymbolDropTable `json:"symbolDT"`
	Multiplier      int             `json:"multiplier"`
	LastPlayed      time.Time       `json:"lastPlayed"`
	GameVersion     string          `json:"gameVesion"`
	gonnaWin        bool
	windowWidth     int
	windowHeight    int
}

type SymbolDropTable map[emoji.Emoji]int
type TickMsg time.Time
type DistributeTokenMsg bool
type Symbol emoji.Emoji
type Slot [3]Symbol

func initTable() SymbolDropTable {
	return map[emoji.Emoji]int{
		emoji.Keycap1:       0,
		emoji.Keycap2:       0,
		emoji.Keycap3:       100,
		emoji.Keycap4:       0,
		emoji.Keycap5:       100,
		emoji.Keycap6:       0,
		emoji.Keycap7:       100,
		emoji.Keycap8:       0,
		emoji.Keycap9:       0,
		emoji.Keycap10:      100,
		emoji.HoneyPot:      50,
		emoji.Joker:         50,
		emoji.HundredPoints: 50,
		emoji.GameDie:       50,
		emoji.FreeButton:    50,
		emoji.FullMoon:      50,
		emoji.Fire:          50,
		emoji.Sunrise:       0,
		emoji.Skull:         0,
		emoji.Bomb:          0,
	}
}

// roll a number between [min,max)
func roll(min, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	roll := r.Intn(max-min) + min
	return roll
}

func (s SymbolDropTable) rollTable() Symbol {
	maxRoll := 0
	for _, weight := range s {
		maxRoll += weight
	}

	roll := roll(0, maxRoll)
	cumulativeProbability := 0

	for e, prob := range s {
		cumulativeProbability += prob
		if roll < cumulativeProbability {
			return Symbol(e)
		}
	}
	// This should never happen if probabilities sum to 1, but just in case:
	return Symbol(emoji.WhiteFlag)
}

func (m *model) turnToDay() {
	m.IsDay = true
	m.SymbolDropTable[emoji.Sunrise] = 0
	m.SymbolDropTable[emoji.FullMoon] = initTable()[emoji.FullMoon]
	m.SymbolDropTable[emoji.Joker] = initTable()[emoji.Joker]
	m.SymbolDropTable[emoji.Skull] = initTable()[emoji.Skull]
}

func (m *model) turnToNight() {
	m.IsDay = false
	m.SymbolDropTable[emoji.Sunrise] = initTable()[emoji.FullMoon]
	m.SymbolDropTable[emoji.FullMoon] = 0
	m.SymbolDropTable[emoji.Joker] = initTable()[emoji.Joker] * 4
	m.SymbolDropTable[emoji.Skull] = initTable()[emoji.Skull] * 2
}

func (s Symbol) toEmoji() emoji.Emoji {
	return emoji.Emoji(s)
}

func (s Symbol) toString() string {
	emojiString := emoji.Emoji(s).String()
	emojiWidth := uniseg.StringWidth(emojiString)
	maxWidth := 5

	// need to do string magic here because every emoji is not the same width
	if emojiWidth < maxWidth {
		return emojiString + strings.Repeat(" ", maxWidth-emojiWidth)
	}

	return emojiString
}

func (s Symbol) toInt() int {
	switch s.toEmoji() {
	case emoji.Keycap1:
		return 1
	case emoji.Keycap2:
		return 2
	case emoji.Keycap3:
		return 3
	case emoji.Keycap4:
		return 4
	case emoji.Keycap5:
		return 5
	case emoji.Keycap6:
		return 6
	case emoji.Keycap7:
		return 7
	case emoji.Keycap8:
		return 8
	case emoji.Keycap9:
		return 9
	case emoji.Keycap10:
		return 10
	}
	return -1
}

func (s Slot) getFirstNonJokerSymbol() Symbol {
	if s[0] != Symbol(emoji.Joker) {
		return s[0]
	}
	if s[1] != Symbol(emoji.Joker) {
		return s[1]
	}
	if s[2] != Symbol(emoji.Joker) {
		return s[2]
	}
	return Symbol(emoji.Joker)
}

func (s Slot) allSymbolsMatch() bool {
	if s[0] == s[1] && s[0] == s[2] {
		return true
	}
	return false
}

// either returns the number they match if they all match or -1 if they don't match
func (s Slot) allNumbersMatch() int {
	// handle joker logic

	// special case if all symbols are jokers
	if s[0] == Symbol(emoji.Joker) && s[1] == Symbol(emoji.Joker) && s[2] == Symbol(emoji.Joker) {
		return 0
	}

	if (s[0] == Symbol(emoji.Joker) && s[1] == s[2]) ||
		(s[1] == Symbol(emoji.Joker) && s[0] == s[2]) ||
		(s[2] == Symbol(emoji.Joker) && s[0] == s[1]) {
		return s.getFirstNonJokerSymbol().toInt()
	}

	if s[0] == Symbol(emoji.Joker) && s[1] == Symbol(emoji.Joker) ||
		s[1] == Symbol(emoji.Joker) && s[2] == Symbol(emoji.Joker) ||
		s[0] == Symbol(emoji.Joker) && s[2] == Symbol(emoji.Joker) {
		return s.getFirstNonJokerSymbol().toInt()
	}

	// no jokers here :)
	if s[0] == s[1] && s[0] == s[2] {
		return s[0].toInt()
	}

	// no match
	return -1
}

func (s Slot) toStringArray() []string {
	strs := []string{}
	for i := 0; i < 3; i++ {
		strs = append(strs, s[i].toString())
	}
	return strs
}

func initSlot() Slot {
	slot := Slot{}
	for i := 0; i < 3; i++ {
		slot[i] = Symbol(emoji.SlotMachine)
	}
	return slot
}

func (m *model) SpinSlots() {
	if m.TicksLeft > 60 {
		m.Slot[0] = m.SymbolDropTable.rollTable()
	}
	if m.TicksLeft > 30 {
		m.Slot[1] = m.SymbolDropTable.rollTable()
	}
	m.Slot[2] = m.SymbolDropTable.rollTable()

	if m.gonnaWin {
		if m.TicksLeft == 31 {
			m.Slot[1] = m.Slot[0]
			return
		}
		if m.TicksLeft == 1 {
			m.Slot[2] = m.Slot[0]
			return
		}
	}
}

func (m *model) handleWin() {
	if m.Slot[0] == Symbol(emoji.Joker) &&
		m.Slot[1] == Symbol(emoji.Joker) &&
		m.Slot[2] == Symbol(emoji.Joker) {
		m.SpinnerMsg = "Hahahahahahahaha " +
			emoji.RollingOnTheFloorLaughing.String() +
			emoji.RollingOnTheFloorLaughing.String() +
			emoji.RollingOnTheFloorLaughing.String()
		return
	}

	switch m.Slot.getFirstNonJokerSymbol() {
	case Symbol(emoji.HoneyPot):
		m.Dollars += m.JackpotAmount
		m.SpinnerMsg = "🎉🎉🎉YOU HIT THE JACKPOT!!! YOU WON " + strconv.Itoa(
			m.JackpotAmount,
		) + " dollars!!!!🎉🎉🎉"
		m.JackpotAmount = 0
	case Symbol(emoji.HundredPoints):
		m.Dollars += 100
		m.SpinnerMsg = "🎉 Nice 100! You won 100 dollars!"
	case Symbol(emoji.GameDie):
		roll := roll(1, 6)
		m.Multiplier = roll
		m.SpinnerMsg = "Your new multiplier is x" + strconv.Itoa(roll)
	case Symbol(emoji.FreeButton):
		m.IsFreeSpin = true
		m.SpinnerMsg = "Nice! You got a free spin! Go ahead and reroll!"
	case Symbol(emoji.Sunrise):
		m.turnToDay()
		m.SpinnerMsg = "It's back to being daytime! The chances for jokers and skulls is back to normal."
	case Symbol(emoji.FullMoon):
		m.turnToNight()
		m.SpinnerMsg = "It's nighttime now! The chance of jokers and skulls is increased!"
	case Symbol(emoji.Fire):
		m.FeverMode = true
		m.SpinnerMsg = "You're in fever mode now!"
	case Symbol(emoji.Skull):
		roll := roll(1, m.Dollars)
		m.Dollars -= roll
		m.SpinnerMsg = emoji.Skull.String() +
			"You lost " + strconv.Itoa(roll) + " dollars!!!" +
			emoji.Skull.String()
	case Symbol(emoji.Keycap1):
		fallthrough
	case Symbol(emoji.Keycap2):
		fallthrough
	case Symbol(emoji.Keycap3):
		fallthrough
	case Symbol(emoji.Keycap4):
		fallthrough
	case Symbol(emoji.Keycap5):
		fallthrough
	case Symbol(emoji.Keycap6):
		fallthrough
	case Symbol(emoji.Keycap7):
		fallthrough
	case Symbol(emoji.Keycap8):
		fallthrough
	case Symbol(emoji.Keycap9):
		fallthrough
	case Symbol(emoji.Keycap10):
		feverMulti := 1
		feverTxt := ""
		if m.FeverMode {
			m.FeverMode = false
			feverMulti = 2
			feverTxt = "\nBut doubled because you were in fever mode!" + emoji.Fire.String()

		}
		baseWonDollars := m.Slot.getFirstNonJokerSymbol().toInt() * m.Multiplier
		finalWonDollars := baseWonDollars * feverMulti
		m.Dollars += finalWonDollars
		m.SpinnerMsg = "🎉You won " + strconv.Itoa(baseWonDollars) + " dollars!🎉" + feverTxt
	}
}

func (m model) saveGame() {
	m.updateLastPlayed()
	data, err := json.Marshal(m)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
	err = os.WriteFile(SAVE_NAME, data, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}

func (m *model) finishSpin() {
	m.Spinning = false
	if m.Slot.allSymbolsMatch() || m.Slot.allNumbersMatch() != -1 {
		m.handleWin()
	} else {
		m.SpinnerMsg = "Try again :("
		m.JackpotAmount += LOSE_JACKPOT_INCREASE
	}
	m.gonnaWin = false
	m.saveGame()
}

func (m *model) doTick() tea.Cmd {
	m.SpinSlots()
	m.TicksLeft--
	if m.TicksLeft == 0 {
		m.finishSpin()
		return nil
	}
	return tea.Tick(time.Second/40, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m *model) updateLastPlayed() {
}

func initialGame() model {
	return model{
		Tokens:          100,
		Dollars:         0,
		TicksLeft:       0,
		Spinning:        false,
		SpinnerMsg:      "Spin to win!",
		IsDay:           true,
		JackpotAmount:   0,
		Multiplier:      1,
		IsFreeSpin:      false,
		FeverMode:       false,
		Slot:            initSlot(),
		SymbolDropTable: initTable(),
		LastPlayed:      time.Now(),
		GameVersion:     GAME_VERSION,
	}
}

func loadGame() model {
	jsonData, err := os.ReadFile(SAVE_NAME)
	if err != nil {
		return initialGame()
	}

	var model model
	err = json.Unmarshal(jsonData, &model)
	if err != nil {
		panic("Error unmarshalling JSON: " + err.Error())
	}

	// if the game gets upgraded, reset the local game
	// TODO migrate data lol
	if model.GameVersion != GAME_VERSION {
		return initialGame()
	}

	return model
}

func (m *model) distributeTokens() {
	m.Tokens += DAILY_TOKENS
	m.SpinnerMsg = "It's a new day! Have " +
		strconv.Itoa(DAILY_TOKENS) +
		" tokens on the house :)"
	m.saveGame()
}

func (m model) Init() tea.Cmd {
	now := time.Now()
	if m.LastPlayed.Year() < now.Year() ||
		m.LastPlayed.Month() < now.Month() ||
		m.LastPlayed.Day() < now.Day() {
		return tea.Cmd(func() tea.Msg { return DistributeTokenMsg(true) })
	}
	if m.Spinning {
		return m.doTick()
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

	case TickMsg:
		return m, m.doTick()

	case DistributeTokenMsg:
		m.distributeTokens()
		return m, nil

	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q":
			m.saveGame()
			return m, tea.Quit

		case "x":
			if m.Dollars < TOKEN_COST {
				m.SpinnerMsg = "You don't have enough money to purchase those tokens!"
				return m, nil
			}
			m.Tokens += 10
			m.Dollars -= TOKEN_COST
			m.SpinnerMsg = "You bought 10 tokens for $" + strconv.Itoa(TOKEN_COST) + "!"
			return m, nil

		case " ":
			if m.Spinning {
				m.SpinnerMsg = "Chill, the spinner is still spinning!"
				return m, nil
			} else if m.Tokens == 0 && !m.IsFreeSpin {
				m.SpinnerMsg = "You have no more tokens!"
				return m, nil
			} else if m.Tokens < 0 {
				m.SpinnerMsg = "How the heck do you have negative tokens?"
				return m, nil
			}

			// free spin logic
			if m.IsFreeSpin {
				m.SpinnerMsg = "Enjoy the free spin!"
				m.IsFreeSpin = false
			} else {
				m.SpinnerMsg = ""
				m.Tokens--
			}
			r := roll(0, 100)
			if r < CHANCE_TO_WIN {
				m.gonnaWin = true
			}
			m.TicksLeft = SPIN_TICKS
			m.Spinning = true
			m.saveGame()
			return m, m.doTick()
		}
	}

	return m, nil
}

func (m model) View() string {
	slotStr := "         " + lipgloss.NewStyle().
		Render(m.Slot.toStringArray()...) +
		"  x" + strconv.Itoa(m.Multiplier) + " multiplier"

	tokenStr := strconv.Itoa(m.Tokens)
	dollarsStr := strconv.Itoa(m.Dollars)
	jackpotAmountStr := strconv.Itoa(m.JackpotAmount)

	feverModeStr := ""
	if m.FeverMode {
		feverModeStr = "\n" +
			emoji.Fire.String() +
			"You're on fire! Your next win will be doubled!" +
			emoji.Fire.String()
	}

	timeStr := "It is currently daytime! You are safe " + emoji.Sunrise.String()
	if !m.IsDay {
		timeStr = "It is currently nighttime! Watch out for skulls " + emoji.Skull.String()
	}

	consoleTxt := "\n\n" + m.SpinnerMsg +
		"\nPress the spacebar to spin!" +
		"\nYou have " + tokenStr + " tokens left" +
		"\nYou have " + dollarsStr + " dollars" +
		"\n" + timeStr +
		"\nThe current jackpot is worth " + jackpotAmountStr + " dollars!" +
		feverModeStr +
		"\nPress 'q' to quit" +
		"\n\nShop" +
		"\n Press 'x' to exchange $" + strconv.Itoa(TOKEN_COST) + " for 10 tokens" +
		"\n"

	console := lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).Height(10)

	style := lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center)

	screenTxt := lipgloss.JoinVertical(
		lipgloss.Center,
		style.Render(slotStr),
		console.Render(consoleTxt),
	)

	screen := lipgloss.NewStyle().
		Width(m.windowWidth-2).
		Height(m.windowHeight-2).
		Align(lipgloss.Center, lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("2")).
		Render(screenTxt)

	return screen
}

func main() {
	p := tea.NewProgram(loadGame(), tea.WithFPS(120), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
