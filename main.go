package main

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type planetNum int

type fastSeed struct {
	a uint8
	b uint8
	c uint8
	d uint8
}

type seed struct {
	w0 uint16
	w1 uint16
	w2 uint16
}

type planSys struct {
	x            uint
	y            uint /* One byte unsigned */
	economy      uint /* These two are actually only 0-7  */
	govType      uint
	techLev      uint /* 0-16 i think */
	population   uint /* One byte */
	productivity uint /* Two byte */
	radius       uint /* Two byte (not used by game at all) */
	goatSoupSeed fastSeed
	name         string
	description  string
}

const galSize = 256
const alienItems = 16
const lastTrade = alienItems
const maxLen = 20

const UNIT_TONNES = 0
const UNIT_KG = 1
const UNIT_G = 2

const numForLave = 7 /* Lave is 7th generated planet in galaxy one */
const numforZaonce = 129
const numforDiso = 147
const numforRied = 46

//var seed seedType
var rndSeed fastSeed

var galSeed seed

type tradeGood struct {
	/* In 6502 version these were: */
	basePrice uint   /* one byte */
	gradient  int16  /* five bits plus sign */
	baseQuant uint   /* one byte */
	maskByte  uint   /* one byte */
	units     uint   /* two bits */
	name      string /* longest="Radioactives" */
}

type market struct {
	quantity [lastTrade + 1]uint
	price    [lastTrade + 1]uint
}

const base0 = 0x5A4A
const base1 = 0x0248
const base2 = 0xB753 /* Base seed for galaxy 1 */

//static const char *digrams=
//							 "ABOUSEITILETSTONLONUTHNO"
//							 "ALLEXEGEZACEBISO"
//							 "USESARMAINDIREA?"
//							 "ERATENBERALAVETI"
//							 "EDORQUANTEISRION";

// 1.5 planet names fix
var pairs0 = "ABOUSEITILETSTONLONUTHNOALLEXEGEZACEBISOUSESARMAINDIREA.ERATENBERALAVETIEDORQUANTEISRION"
var pairs = "..LEXEGEZACEBISO" +
	"USESARMAINDIREA." +
	"ERATENBERALAVETI" +
	"EDORQUANTEISRION" /* Dots should be nullprint characters */

var govNames = []string{"Anarchy", "Feudal", "Multi-gov", "Dictatorship",
	"Communist", "Confederacy", "Democracy", "Corporate State"}

var econNames = []string{"Rich Ind", "Average Ind", "Poor Ind", "Mainly Ind",
	"Mainly Agri", "Rich Agri", "Average Agri", "Poor Agri"}

var unitNames = []string{"t", "kg", "g"}

const politicallyCorrect = false

/* Player workspace */
// gameState model copied from the javascript port by Joshua Bell: TODO - INSERT LINK

type gameState struct {

	// PRNG
	useNativeRand bool
	lastRand      int

	// Galaxy
	galaxyNum uint
	galaxy    []planSys /* Need 0 to galsize-1 inclusive */

	// Current System

	localMarket   market
	currentPlanet planetNum

	// Ship
	shipsHold []uint
	cash      int32
	fuel      uint
	holdSpace uint
	fuelCost  int
	maxFuel   int
}

type elite struct {
	state gameState
}

func newGameState() gameState {
	var gState gameState

	gState.useNativeRand = false
	gState.lastRand = 0

	gState.galaxyNum = 0 /* Galaxy number (1-8) */
	//gState.galaxy = nil

	gState.currentPlanet = 0
	//gState.localMarket = nil

	gState.cash = 0
	gState.fuel = 0
	gState.holdSpace = 0
	gState.fuelCost = 2 /* 0.2 CR/Light year */
	gState.maxFuel = 70 /* 7.0 LY tank */

	gState.shipsHold = make([]uint, len(commodities()))

	return gState

}

// Implements the politically correct options for the
// commodity names
func commodities() []tradeGood {

	/* Set to true for NES-sanitised trade goods */

	/* Data for DB's price/availability generation system */
	/*                   Base  Grad Base Mask Un   Name
	price ient quant     it              */

	var servants string
	var beverages string
	var drugs string

	if politicallyCorrect {
		servants = "Robot Slaves"
		beverages = "Beverages"
		drugs = "Rare Species"

	} else {
		servants = "Slaves"
		beverages = "Liquor/Wines"
		drugs = "Narcotics"
	}

	initCommodities := []tradeGood{
		tradeGood{basePrice: 0x13,
			gradient:  -0x02,
			baseQuant: 0x06,
			maskByte:  0x01,
			units:     UNIT_TONNES,
			name:      "Food",
		},
		tradeGood{basePrice: 0x14,
			gradient:  -0x01,
			baseQuant: 0x0A,
			maskByte:  0x03,
			units:     UNIT_TONNES,
			name:      "Textiles",
		},
		tradeGood{basePrice: 0x41,
			gradient:  -0x03,
			baseQuant: 0x02,
			maskByte:  0x07,
			units:     UNIT_TONNES,
			name:      "Radioactives",
		},
		tradeGood{basePrice: 0x28,
			gradient:  -0x05,
			baseQuant: 0xE2,
			maskByte:  0x1F,
			units:     UNIT_TONNES,
			name:      servants,
		},
		tradeGood{basePrice: 0x53,
			gradient:  -0x05,
			baseQuant: 0xFB,
			maskByte:  0x0F,
			units:     UNIT_TONNES,
			name:      beverages,
		},
		tradeGood{basePrice: 0xE4,
			gradient:  +0x08,
			baseQuant: 0x36,
			maskByte:  0x03,
			units:     UNIT_TONNES,
			name:      "Luxeries",
		},
		tradeGood{basePrice: 0xEB,
			gradient:  +0x1D,
			baseQuant: 0x08,
			maskByte:  0x78,
			units:     UNIT_TONNES,
			name:      drugs,
		},
		tradeGood{basePrice: 0x9A,
			gradient:  +0x0E,
			baseQuant: 0x38,
			maskByte:  0x03,
			units:     UNIT_TONNES,
			name:      "Computers",
		},
		tradeGood{basePrice: 0x75,
			gradient:  +0x06,
			baseQuant: 0x28,
			maskByte:  0x07,
			units:     UNIT_TONNES,
			name:      "Machinery",
		},
		tradeGood{basePrice: 0x4E,
			gradient:  +0x01,
			baseQuant: 0x11,
			maskByte:  0x1F,
			units:     UNIT_TONNES,
			name:      "Alloys",
		},
		tradeGood{basePrice: 0x7C,
			gradient:  +0x0D,
			baseQuant: 0x1D,
			maskByte:  0x07,
			units:     UNIT_TONNES,
			name:      "Firearms",
		},
		tradeGood{basePrice: 0xB0,
			gradient:  -0x09,
			baseQuant: 0xDC,
			maskByte:  0x3F,
			units:     UNIT_TONNES,
			name:      "Furs",
		},
		tradeGood{basePrice: 0x20,
			gradient:  -0x01,
			baseQuant: 0x35,
			maskByte:  0x03,
			units:     UNIT_TONNES,
			name:      "Minerals",
		},
		tradeGood{basePrice: 0x61,
			gradient:  -0x01,
			baseQuant: 0x42,
			maskByte:  0x07,
			units:     UNIT_KG,
			name:      "Gold",
		},
		tradeGood{basePrice: 0xAB,
			gradient:  -0x02,
			baseQuant: 0x37,
			maskByte:  0x1F,
			units:     UNIT_KG,
			name:      "Platinum",
		},
		tradeGood{basePrice: 0x2D,
			gradient:  -0x01,
			baseQuant: 0xFA,
			maskByte:  0x0F,
			units:     UNIT_G,
			name:      "Gem-Stones",
		},
		tradeGood{basePrice: 0x35,
			gradient:  +0x0F,
			baseQuant: 0xC0,
			maskByte:  0x07,
			units:     UNIT_TONNES,
			name:      "Alien Items",
		},
	}

	return initCommodities
}

/**-Required data for text interface **/

var tradeNames [lastTrade]string /* Tradegood names used in text commands. Set using commodities array */

func goodsNames() []string {

	var names []string
	commoditiesList := commodities()

	names = make([]string, len(commoditiesList))

	for count, tg := range commoditiesList {
		names[count] = tg.name
	}

	return names
}

func newMarket(fluct uint, p planSys) market {
	/* Prices and availabilities are influenced by the planet's economy type
	   (0-7) and a random "fluctuation" byte that was kept within the saved
	   commander position to keep the market prices constant over gamesaves.
	   Availabilities must be saved with the game since the player alters them
	   by buying (and selling(?))

	   Almost all operations are one byte only and overflow "errors" are
	   extremely frequent and exploited.

	   Trade Item prices are held internally in a single byte=true value/4.
	   The decimal point in prices is introduced only when printing them.
	   Internally, all prices are integers.
	   The player's cash is held in four bytes.
	*/

	var market market

	marketCommodities := commodities()

	for i := 0; i <= lastTrade; i++ {
		//int q
		product := int16((p.economy)) * (marketCommodities[i].gradient)
		changing := fluct & (marketCommodities[i].maskByte)

		q := uint16((marketCommodities[i].baseQuant)) + uint16(changing) - uint16(product)
		q = q & 0xFF

		if q&0x80 == 1 {
			q = 0
		} /* Clip to positive 8-bit */

		market.quantity[i] = uint((q & 0x3F)) /* Mask to 6 bits */

		q = uint16((marketCommodities[i].basePrice) + changing + uint(product))
		q = q & 0xFF
		market.price[i] = uint(q * 4)
	}

	market.quantity[alienItems] = 0 /* Override to force nonavailability */
	return market
}

func (gs *gameState) displaymarket() bool {

	marketCommodities := commodities()

	for i := 0; i <= lastTrade; i++ {
		fmt.Printf("\n")

		if len(marketCommodities[i].name) <= 6 {
			fmt.Printf("%s\t", marketCommodities[i].name)
		} else {
			fmt.Printf("%s", marketCommodities[i].name)
		}

		fmt.Printf("\t%.1f", float32((gs.localMarket.price[i]))/10)
		fmt.Printf("\t%d", gs.localMarket.quantity[i])
		fmt.Printf("%s", unitNames[marketCommodities[i].units])
		fmt.Printf("\t%d", gs.shipsHold[i])
	}

	fmt.Printf("\n\nFuel: %.1f", float32(gs.fuel)/10)
	fmt.Printf("\tHoldspace: %dt", gs.holdSpace)

	return true
}

// COMMANDS

// All valid commands
var validCommands = []string{"buy", "sell", "fuel", "jump", "cash", "mkt", "help", "hold", "sneak", "local", "info", "galhyp", "quit", "rand"}

// A map of command names to the functions that perform the command. Since most commands alter the game state, this is part of the
// gamestate logic, as are the commands.

func (gs *gameState) commands() map[string]interface{} {

	commandmap := map[string]interface{}{
		"buy":    gs.buy,
		"hold":   gs.hold,
		"sell":   gs.sell,
		"fuel":   gs.manageFuel,
		"jump":   gs.jump,
		"cash":   gs.manageCash,
		"mkt":    gs.displaymarket,
		"help":   gs.help,
		"sneak":  gs.sneak,
		"local":  gs.local,
		"galhyp": gs.galhyp,
		"info":   gs.info,
		"quit":   gs.quit,
		"rand":   gs.useWeakRand,
	}

	return commandmap
}

func (gs *gameState) help() bool {
	fmt.Printf("\nCommands are:")
	fmt.Printf("\nBuy   tradegood ammount")
	fmt.Printf("\nSell  tradegood ammount")
	fmt.Printf("\nFuel  ammount    (buy ammount LY of fuel)")
	fmt.Printf("\nJump  planetname (limited by fuel)")
	fmt.Printf("\nSneak planetname (any distance - no fuel cost)")
	fmt.Printf("\nGalhyp           (jumps to next galaxy)")
	fmt.Printf("\nInfo  planetname (prints info on system")
	fmt.Printf("\nMkt              (shows market prices)")
	fmt.Printf("\nLocal            (lists systems within 7 light years)")
	fmt.Printf("\nCash number      (alters cash - cheating!)")
	fmt.Printf("\nHold number      (change cargo bay)")
	fmt.Printf("\nQuit or ^C       (exit)")
	fmt.Printf("\nHelp             (display this text)")
	fmt.Printf("\nRand             (toggle RNG)")
	fmt.Printf("\n\nAbbreviations allowed eg. b fo 5 = Buy Food 5, m= Mkt")

	return true
}

func (gs *gameState) local() bool {

	fmt.Printf("Galaxy number %d", gs.galaxyNum)

	for syscount := 0; syscount < galSize; syscount++ {
		d := distance(gs.galaxy[syscount], gs.galaxy[gs.currentPlanet])
		if d <= uint(gs.maxFuel) {
			if d <= gs.fuel {
				fmt.Printf("\n * ")
			} else {
				fmt.Printf("\n - ")
			}

			printSys(gs.galaxy[syscount], true)

			fmt.Printf(" (%.1f LY)", (float64(d) / float64(10)))
		}
	}

	return true
}

func (gs *gameState) manageFuel(ammount string) bool {

	x, _ := strconv.Atoi(string(ammount))

	f := uint(x) * 10

	if f+gs.fuel > uint(gs.maxFuel) {
		f = uint(gs.maxFuel) - gs.fuel
	}

	if gs.fuelCost > 0 {
		fmt.Printf("Cash: %d\n", gs.cash)
		fmt.Printf("Fuel Cost: %d\n", gs.fuelCost)

		if int32(f)*int32(gs.fuelCost) > gs.cash {

			f = uint(gs.cash / int32(gs.fuelCost))
		}
	}

	fmt.Printf("F: %d", f)
	gs.fuel += f
	gs.cash -= int32(gs.fuelCost) * int32(f)

	if f == 0 {
		fmt.Println("\n Can't buy any fuel")
	} else {
		fmt.Printf("\nBuying %.1fLY fuel", float64(f)/float64(10))
	}

	return true
}

func (gs *gameState) manageCash(transaction string) bool {

	var ammount float64
	op := string(transaction[0])

	if op != "+" && op != "-" {
		fmt.Println("Invlaid Operation")
		return false
	}

	tmpFloat, ferr := strconv.ParseFloat(string(transaction[1:]), 64)

	if ferr == nil {
		ammount = math.Floor(10 * tmpFloat)
	} else {
		ammount = 0
	}

	if op == "+" {
		gs.cash += int32(ammount)
	}

	if op == "-" {
		gs.cash -= int32(ammount)
	}

	return true
}

func (gs *gameState) buyFuel(ammount string) bool /* Attempt to buy f tonnes of fuel */ {

	return true
}

func (gs *gameState) quit() {
	os.Exit(0)
}

func (gs *gameState) useWeakRand() {
	gs.useNativeRand = !gs.useNativeRand
}

// Implementations for each command

// Inputs will be strings because they come from the command line
func (gs *gameState) buy(good string, amount string) bool {
	var t uint

	amountToBuy, err := strconv.Atoi(amount)
	baseCommodities := commodities()

	if err != nil || amountToBuy == 0 {
		amountToBuy = 1
	}

	fmt.Println(amountToBuy)

	isGood, goodIdx := strInArray(goodsNames(), good)

	if !isGood {
		fmt.Println("Unknown trade good")
		return false
	}

	uGood := uint(goodIdx)
	uAmount := uint(amountToBuy)

	if gs.cash < 0 {
		t = 0

	} else {
		// could use math.Min here.
		t = myMin(gs.localMarket.quantity[uGood], uAmount)

		if (baseCommodities[uGood].units) == UNIT_TONNES {
			t = myMin(gs.holdSpace, t)
		}

		t = myMin(t, uint(math.Floor(float64(gs.cash/int32(gs.localMarket.price[uGood])))))
	}

	gs.shipsHold[uGood] += t
	gs.localMarket.quantity[uGood] -= t
	gs.cash -= int32(t) * int32(gs.localMarket.price[uGood])

	if baseCommodities[uGood].units == UNIT_TONNES {
		gs.holdSpace -= t
	}

	if t == 0 {
		fmt.Println("Cannot buy anything")
		return false
	}

	fmt.Println("Buying " + amount + unitNames[baseCommodities[goodIdx].units] + " of " + good)

	return true
}

func (gs *gameState) sell(good string, amount string) bool {
	var t uint

	amountToSell, err := strconv.Atoi(amount)
	baseCommodities := commodities()

	if err != nil || amountToSell == 0 {
		amountToSell = 1

	} else {
		return false
	}

	isGood, goodIdx := strInArray(goodsNames(), good)

	if !isGood {
		fmt.Println("Unknown trade good")
		return false
	}

	uGood := uint(goodIdx)
	uAmount := uint(amountToSell)

	// could use math.Min here.
	t = myMin(gs.shipsHold[uGood], uAmount)
	gs.shipsHold[uGood] -= t
	gs.localMarket.quantity[uGood] += t
	gs.cash += int32(t) * int32(gs.localMarket.price[uGood])

	if baseCommodities[uGood].units == UNIT_TONNES {
		gs.holdSpace += t
	}

	if t == 0 {
		fmt.Println("Cannot sell anything")
		return false
	}

	fmt.Println("Selling " + amount + unitNames[baseCommodities[goodIdx].units] + " of " + good)

	return true
}

func (gs *gameState) jump(s string) bool {

	var d uint

	dest := gs.matchSys(s)

	if dest == gs.currentPlanet {
		fmt.Println("\nBad jump")
		return false
	}

	d = distance(gs.galaxy[dest], gs.galaxy[gs.currentPlanet])

	if d > gs.fuel {
		fmt.Println("\nJump to far")
		return false
	}

	gs.fuel -= d

	gs.currentPlanet = dest
	gs.localMarket = newMarket(gs.randByte(), gs.galaxy[dest])

	printSys(gs.galaxy[gs.currentPlanet], false)

	return true
}

func (gs *gameState) sneak(s string) bool {

	currentFuel := gs.fuel
	gs.fuel = 666

	b := gs.jump(s)

	gs.fuel = currentFuel

	return b
}

/* Preserve planetnum (eg. if leave 7th planet
   arrive at 7th planet)
   Classic Elite always jumped to planet nearest (0x60,0x60)
*/
func (gs *gameState) galhyp() bool { /* Jump to next galaxy */

	gs.galaxyNum++

	if gs.galaxyNum == 9 {
		gs.galaxyNum = 1
	}

	gs.galaxy = buildGalaxy(gs.galaxyNum)

	return true
}

// Uses reflection to call the function specified in the passed map.
func doCmd(m map[string]interface{}, name string, params ...interface{}) (result []reflect.Value, err error) {
	f := reflect.ValueOf(m[name])
	fmt.Println(name)
	if len(params) != f.Type().NumIn() {
		err = errors.New("The number of params is not adapted")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	return
}

func (gs *gameState) matchSys(s string) planetNum {
	/* Return id of the planet whose name matches passed strinmg
	   closest to currentplanet - if none return currentplanet */
	//var sysCount planetNum
	var p planetNum
	var sysCount planetNum

	p = gs.currentPlanet

	var d uint = 9999

	/*
		{	if (distance(galaxy[syscount],galaxy[currentplanet])<d)
			{ d=distance(galaxy[syscount],galaxy[currentplanet]);
			  p=syscount;
			}
		  }
	 	} */

	for sysCount = 0; sysCount < galSize; sysCount++ {
		if strings.HasPrefix(strings.ToUpper(s), gs.galaxy[sysCount].name) {
			if distance(gs.galaxy[sysCount], gs.galaxy[gs.currentPlanet]) < d {

				d = distance(gs.galaxy[sysCount], gs.galaxy[gs.currentPlanet])
				p = sysCount
			}
		}
	}

	return p
}

func distance(a planSys, b planSys) uint {
	/* Seperation between two planets (4*sqrt(X*X+Y*Y/4)) */

	// return (uint)ftoi(4*sqrt((a.x-b.x)*(a.x-b.x)+(a.y-b.y)*(a.y-b.y)/4));
	ax := float64(a.x)
	ay := float64(a.y)

	bx := float64(b.x)
	by := float64(b.y)
	d := 4 * math.Sqrt((ax-bx)*(ax-bx)+(ay-by)*(ay-by)/4)

	return uint(d)
}

func (gs *gameState) info(s string) bool {

	dest := gs.matchSys(s)
	printSys(gs.galaxy[dest], false)

	return true
}

func (gs *gameState) hold(s string) bool {

	a, err := strconv.Atoi(s)

	if err != nil {
		return false
	}

	var t uint

	allCommodities := commodities()

	for i := 0; i < len(allCommodities); i++ {

		if allCommodities[i].units == UNIT_TONNES {
			t += gs.shipsHold[i]
		}
	}

	if t > uint(a) {
		fmt.Printf("\nHold to full")
		return false
	}

	gs.holdSpace = uint(a) - t

	return true
}

// Obey command s
func (gs *gameState) parser(s string) bool {

	var reflectRet []reflect.Value
	var err error

	s = strings.ToLower(s)
	cmdArr := strings.Split(s, " ")

	cmd := cmdArr[0]
	var args []string

	if len(cmd) > 1 {
		args = cmdArr[1:]
	}

	isValid, _ := strInArray(validCommands, cmd)
	if isValid {

		// Probably a better way of doing this!

		if len(args) == 0 {
			reflectRet, err = doCmd(gs.commands(), cmd)

		} else if len(args) == 1 {
			reflectRet, err = doCmd(gs.commands(), cmd, args[0])

		} else if len(args) == 2 {
			reflectRet, err = doCmd(gs.commands(), cmd, args[0], args[1])

		} else if len(args) == 3 {
			reflectRet, err = doCmd(gs.commands(), cmd, args[0], args[1], args[2])

		} else if len(args) == 4 {
			reflectRet, err = doCmd(gs.commands(), cmd, args[0], args[1], args[2], args[3])

		} else {
			fmt.Println("Wrong number of arguments")
		}

	} else {
		fmt.Println("Unknown Command")
		return false
	}

	if err != nil {
		fmt.Println("Command returned an error: " + err.Error())
		return false
	}

	return reflectRet[0].Bool()
}

func (gs *gameState) mySRand(seed int) {
	gs.lastRand = seed - 1
}

func (gs *gameState) myRand() int {
	var r int

	if gs.useNativeRand {
		r = rand.Int()
	} else {
		// As supplied by D McDonnell	from SAS Insititute C
		r = (((((((((((gs.lastRand << 3) - gs.lastRand) << 3) + gs.lastRand) << 1) + gs.lastRand) << 4) - gs.lastRand) << 1) - gs.lastRand) + 0xe60) & 0x7fffffff
		gs.lastRand = r - 1
	}
	return r
}

func myMin(a uint, b uint) uint {
	if a < b {
		return a
	}

	return b
}

// PRNG

func (gs *gameState) randByte() uint {
	return uint(gs.myRand()) & 0xFF
}

func newFastSeed(a, b, c, d uint8) fastSeed {
	var fs fastSeed

	fs.a = a
	fs.b = b
	fs.c = c
	fs.d = d

	return fs
}

func (fs *fastSeed) next() int {

	var a int

	//fmt.Printf("\nseed a: %d seed b: %d seed c: %d seed d: %d\n", fs.a, fs.b, fs.c, fs.d)

	x := (fs.a * 2) & 0xFF
	//fmt.Printf("X: %d\n", x)
	//fmt.Printf("fs.c: %d\n", fs.c)
	//fmt.Printf("x + c: %d\n", int(x)+int(fs.c))

	a = int(x) + int(fs.c)
	//fmt.Printf("a1: %d\n", a)

	if fs.a > 127 {
		fmt.Printf("A is %d\n", fs.a)
		a++
	}

	//fs.a = uint8(a & 0xFF)
	fs.a = uint8(a) & 0xFF
	fs.c = x

	a = a >> 8 // a = any carry left from above
	//fmt.Printf("a2: %d\n", a)

	x = fs.b
	a = (a + int(x) + int(fs.d)) & 0xFF
	//fmt.Printf("a3: %d\n", a)
	fs.b = uint8(a)
	fs.d = x

	//fmt.Printf("RND NO: %d\n", a)
	return a
}

/*


var x = (this.a << 1) & 0xFF;
        var a = x + this.c;
        if (this.a > 127) {
            a += 1;
        }
        this.a = a & 0xFF;
        this.c = x;

        a = a >> 8;
        x = this.b;
        a = (a + x + this.d) & 0xFF;
        this.b = a;
        this.d = x;
		return a;



int gen_rnd_number (void)
{	int a,x;
	x = (rnd_seed.a * 2) & 0xFF;
	a = x + rnd_seed.c;
	if (rnd_seed.a > 127)	a++;
	rnd_seed.a = a & 0xFF;
	rnd_seed.c = x;

	a = a / 256;	// a = any carry left from above
	x = rnd_seed.b;
	a = (a + x + rnd_seed.d) & 0xFF;
	rnd_seed.b = a;
	rnd_seed.d = x;

  printf("\nRND NO: %u\n", a);
	return a;
}
*/

func newSeed(w0, w1, w2 uint16) seed {
	var newseed seed

	newseed.w0 = w0
	newseed.w1 = w1
	newseed.w2 = w2

	return newseed
}

func (s *seed) tweakSeed() {
	var temp uint16

	temp = ((s.w0) + (s.w1) + (s.w2)) //& 0xFFFF // 2 Byte arithmetic
	s.w0 = s.w1
	s.w1 = s.w2
	s.w2 = temp
}

// Generate Planetary System

func newPlanSys(s *seed) planSys {
	var ps planSys

	longNameFlag := s.w0 & 0x40

	ps.x = uint(s.w1 >> 8)
	ps.y = uint(s.w0 >> 8)

	//ps.x = uint(math.Floor(float64(s.w1 >> 8)))
	//ps.y = uint(math.Floor(float64(s.w0 >> 8)))

	ps.govType = uint((s.w1 >> 3) & 7)
	ps.economy = uint((s.w0 >> 8) & 7)

	//ps.govType = uint(math.Floor(float64((s.w1 >> 3) & 7))) // bits 3,4 &5 of w1

	//ps.economy = uint(math.Floor((float64((s.w0 >> 8) & 7)))) // bits 8,9 &A of w0

	if ps.govType <= 1 {
		ps.economy = (ps.economy | 2)
	}

	ps.techLev = uint((s.w1>>8)&0x03) + uint(ps.economy^0x07)
	//ps.techLev = uint(math.Floor(float64(((s.w1 >> 8) & 0x03)) + float64(ps.economy^0x07)))

	ps.techLev += ps.govType >> 1

	if ps.govType&0x01 == 1 {
		/* C simulation of 6502's LSR then ADC */
		ps.techLev++
	}

	ps.population = 4*ps.techLev + ps.economy
	ps.population += ps.govType + 1

	ps.productivity = ((ps.economy ^ 0x07) + 3) * (ps.govType + 4)
	ps.productivity *= ps.population * 8

	ps.radius = uint(256*(((s.w2>>8)&0x0f)+11)) + ps.x

	// Seed for "goat soup" description
	ps.goatSoupSeed = newFastSeed(uint8(s.w1&0xFF), uint8(s.w1>>8), uint8(s.w2&0xFF), uint8(s.w2>>8))

	//fmt.Println(ps.goatSoupSeed)
	// Name

	pair1 := ((s.w2 >> 8) & 0x1F) << 1
	s.tweakSeed()
	pair2 := ((s.w2 >> 8) & 0x1F) << 1
	s.tweakSeed()
	pair3 := ((s.w2 >> 8) & 0x1F) << 1
	s.tweakSeed()
	pair4 := ((s.w2 >> 8) & 0x1F) << 1
	s.tweakSeed()
	/* Always four iterations of random number */

	name := string(pairs[pair1])
	name += string(pairs[pair1+1])
	name += string(pairs[pair2])
	name += string(pairs[pair2+1])
	name += string(pairs[pair3])
	name += string(pairs[pair3+1])

	ps.name = string(pairs[pair1])
	ps.name += string(pairs[pair1+1])
	ps.name += string(pairs[pair2])
	ps.name += string(pairs[pair2+1])
	ps.name += string(pairs[pair3])
	ps.name += string(pairs[pair3+1])

	// TODO: Check this
	if longNameFlag == 1 { /* bit 6 of ORIGINAL w0 flags a four-pair name */
		ps.name += string(pairs[pair4])
		ps.name += string(pairs[pair4+1])
	}

	ps.name = strings.Replace(ps.name, ".", "", -1)
	//fmt.Println("NAME: " + ps.name)

	ps.description = ps.goatSoup("\x8F is \x97.", ps.goatSoupSeed)

	return ps
}

func (ps *planSys) goatSoup(source string, prng fastSeed) string {

	var out string

	desc_list := [][]string{
		/* 81 0*/ {"fabled", "notable", "well known", "famous", "noted"},
		/* 82 1*/ {"very", "mildly", "most", "reasonably", ""},
		/* 83 2*/ {"ancient", "\x95", "great", "vast", "pink"},
		/* 84 3*/ {"\x9E \x9D plantations", "mountains", "\x9C", "\x94 forests", "oceans"},
		/* 85 4*/ {"shyness", "silliness", "mating traditions", "loathing of \x86", "love for \x86"},
		/* 86 5*/ {"food blenders", "tourists", "poetry", "discos", "\x8E"},
		/* 87 6*/ {"talking tree", "crab", "bat", "lobst", "\xB2"},
		/* 88 7*/ {"beset", "plagued", "ravaged", "cursed", "scourged"},
		/* 89 8*/ {"\x96 civil war", "\x9B \x98 \x99s", "a \x9B disease", "\x96 earthquakes", "\x96 solar activity"},
		/* 8A 9*/ {"its \x83 \x84", "the \xB1 \x98 \x99", "its inhabitants' \x9A \x85", "\xA1", "its \x8D \x8E"},
		/* 8B 10*/ {"juice", "brandy", "water", "brew", "gargle blasters"},
		/* 8C 11*/ {"\xB2", "\xB1 \x99", "\xB1 \xB2", "\xB1 \x9B", "\x9B \xB2"},
		/* 8D 12*/ {"fabulous", "exotic", "hoopy", "unusual", "exciting"},
		/* 8E 13*/ {"cuisine", "night life", "casinos", "sit coms", " \xA1 "},
		/* 8F 14*/ {"\xB0", "The planet \xB0", "The world \xB0", "This planet", "This world"},
		/* 90 15*/ {"n unremarkable", " boring", " dull", " tedious", " revolting"},
		/* 91 16*/ {"planet", "world", "place", "little planet", "dump"},
		/* 92 17*/ {"wasp", "moth", "grub", "ant", "\xB2"},
		/* 93 18*/ {"poet", "arts graduate", "yak", "snail", "slug"},
		/* 94 19*/ {"tropical", "dense", "rain", "impenetrable", "exuberant"},
		/* 95 20*/ {"funny", "wierd", "unusual", "strange", "peculiar"},
		/* 96 21*/ {"frequent", "occasional", "unpredictable", "dreadful", "deadly"},
		/* 97 22*/ {"\x82 \x81 for \x8A", "\x82 \x81 for \x8A and \x8A", "\x88 by \x89", "\x82 \x81 for \x8A but \x88 by \x89", "a\x90 \x91"},
		/* 98 23*/ {"\x9B", "mountain", "edible", "tree", "spotted"},
		/* 99 24*/ {"\x9F", "\xA0", "\x87oid", "\x93", "\x92"},
		/* 9A 25*/ {"ancient", "exceptional", "eccentric", "ingrained", "\x95"},
		/* 9B 26*/ {"killer", "deadly", "evil", "lethal", "vicious"},
		/* 9C 27*/ {"parking meters", "dust clouds", "ice bergs", "rock formations", "volcanoes"},
		/* 9D 28*/ {"plant", "tulip", "banana", "corn", "\xB2weed"},
		/* 9E 29*/ {"\xB2", "\xB1 \xB2", "\xB1 \x9B", "inhabitant", "\xB1 \xB2"},
		/* 9F 30*/ {"shrew", "beast", "bison", "snake", "wolf"},
		/* A0 31*/ {"leopard", "cat", "monkey", "goat", "fish"},
		/* A1 32*/ {"\x8C \x8B", "\xB1 \x9F \xA2", "its \x8D \xA0 \xA2", "\xA3 \xA4", "\x8C \x8B"},
		/* A2 33*/ {"meat", "cutlet", "steak", "burgers", "soup"},
		/* A3 34*/ {"ice", "mud", "Zero-G", "vacuum", "\xB1 ultra"},
		/* A4 35*/ {"hockey", "cricket", "karate", "polo", "tennis"},
		/* B0 = <planet name>
		 * B1 = <planet name>ian
		 * B2 = <random name>
		 */
	}

	out = ""

	for {

		if len(source) == 0 {
			break
		}

		c := source[0]
		source = source[1:]
		if strings.Contains(source, "0xB1") {
			fmt.Println("found token b1")
		}

		fmt.Printf("\nC is: %d \n", int(c))

		if c < 0x80 {
			fmt.Println("C is less than 0x80")
			out += fmt.Sprintf("%c", c)

		} else {

			if c <= 0xA4 {

				var rnd = prng.next()

				arg1 := 0
				arg2 := 0
				arg3 := 0
				arg4 := 0

				if rnd >= 0x33 {
					arg1 = 1
				}
				if rnd >= 0x66 {
					arg2 = 1
				}
				if rnd >= 0x99 {
					arg3 = 1
				}
				if rnd >= 0xCC {
					arg4 = 1
				}

				/*
					fmt.Println(desc_list[c-0x81])
					fmt.Printf("a1: %d a2: %d a3: %d a4: %d\n", arg1, arg2, arg3, arg4)
					fmt.Printf("a1+a2+a3+a4: %d\n", arg1+arg2+arg3+arg4)
					fmt.Printf("Selection: %s\n", desc_list[c-0x81][arg1+arg2+arg3+arg4])
				*/

				//fmt.Printf("\nItem: %d Col: %d, String: %s\n", c-0x81, arg1+arg2+arg3+arg4, desc_list[c-0x81][arg1+arg2+arg3+arg4])
				//fmt.Printf("PRNG SEED: %d,%d,%d,%d\n", prng.a, prng.b, prng.c, prng.d)
				//fmt.Printf("rnd: %d\n", rnd)
				fmt.Printf("About to recurse with PRNG values: a=%d, b=%d, c=%d, d=%d\n", prng.a, prng.b, prng.c, prng.d)
				out += ps.goatSoup(desc_list[c-0x81][arg1+arg2+arg3+arg4], prng)
				fmt.Println(out)

			} else {

				switch c {
				case 0xB0: /* planet name */

					out += fmt.Sprintf("%s", strings.ToLower(string(ps.name[0])))

					/*
						for i := 1; i < len(ps.name); i++ {
							out += fmt.Sprintf("%s", strings.ToLower(string(ps.name[i])))
						}
					*/

					for i := 1; i < len(ps.name); i++ {

						out += fmt.Sprintf("%s", strings.ToLower(string(ps.name[i])))
					}
					break

				case 0xB1: /* <planet name>ian */
					out += fmt.Sprintf("%s", strings.ToLower(string(ps.name[0])))
					for i := 1; i < len(ps.name); i++ {
						if (i+1 < len(ps.name)) || ((string(ps.name[i]) != "E") && (string(ps.name[i]) != "I")) {
							out += fmt.Sprintf("%s", strings.ToLower(string(ps.name[i])))
						}
					}
					out += fmt.Sprintf("ian")
					break

				case 0xB2: /* random name */

					len := prng.next() & 3

					for i := 0; int(i) <= len; i++ {

						x := prng.next() & 0x3e
						if i == 0 {
							out += fmt.Sprintf("%s", string(pairs0[x]))
						} else {
							out += fmt.Sprintf("%s", strings.ToLower(string(pairs0[x])))
						}

						out += fmt.Sprintf("%s", strings.ToLower(string(pairs0[x+1])))
					}
					break

				default:
					fmt.Printf("<bad char in data [%X]>", c)
				}
			}
		}
	}
	return out
}

func printSys(plsy planSys, compressed bool) {

	if compressed {

		//	  printf("\n ");
		fmt.Printf("%10s", plsy.name)
		fmt.Printf(" TL: %2d ", (plsy.techLev)+1)
		fmt.Printf("%12s", econNames[plsy.economy])
		fmt.Printf(" %15s", govNames[plsy.govType])
	} else {
		fmt.Printf("\n\nSystem:  ")
		fmt.Printf("%s", plsy.name)
		fmt.Printf("\nPosition (%d,", plsy.x)
		fmt.Printf("%d)", plsy.y)
		fmt.Printf("\nEconomy: (%d) ", plsy.economy)
		fmt.Printf("%s", econNames[plsy.economy])
		fmt.Printf("\nGovernment: (%d) ", plsy.govType)
		fmt.Printf("%s", govNames[plsy.govType])
		fmt.Printf("\nTech Level: %2d", (plsy.techLev)+1)
		fmt.Printf("\nTurnover: %d", (plsy.productivity))
		fmt.Printf("\nRadius: %d", plsy.radius)
		fmt.Printf("\nPopulation: %d Billion", (plsy.population)>>3)

		rndSeed := plsy.goatSoupSeed
		fmt.Printf("\n")
		//fmt.Printf(plsy.description)
		fmt.Printf(plsy.goatSoup("\x8F is \x97.", rndSeed))

	}
}

/* Functions for galactic hyperspace */

func rotatel(x uint16) uint16 { /* rotate 8 bit number leftwards */
	temp := x & 128
	return (2 * (x & 127)) + (temp >> 7)
}

func twist(x uint16) uint16 {
	//return (uint16)((256 * rotatel(x>>8)) + rotatel(x&255))
	return ((256 * rotatel(x>>8)) + rotatel(x&255))
}

func nextgalaxy(s *seed) seed { /* Apply to base seed; once for galaxy 2  */
	var newSeed seed

	newSeed.w0 = twist(s.w0) /* twice for galaxy 3, etc. */
	newSeed.w1 = twist(s.w1) /* Eighth application gives galaxy 1 again*/
	newSeed.w2 = twist(s.w2)

	return newSeed
}

/* Original game generated from scratch each time info needed */
func buildGalaxy(galaxynum uint) []planSys {

	gal := make([]planSys, galSize)

	/* Initialise seed for galaxy 1 */
	galSeed = newSeed(base0, base1, base2)

	for galcount := 1; uint(galcount) < galaxynum; galcount++ {
		galSeed = nextgalaxy(&galSeed)
		//fmt.Println(galSeed)
	}

	/* Put galaxy data into array of structures */
	for syscount := 0; syscount < galSize; syscount++ {
		//galaxy[syscount] = newPlanSys(&galSeed)
		gal[syscount] = newPlanSys(&galSeed)

	}

	return gal
}

// Go Version Utility Functions

func strInArray(strArr []string, itemToCheck string) (bool, int) {
	for idx, item := range strArr {
		if strings.ToLower(item) == strings.ToLower(itemToCheck) {
			return true, idx
		}
	}
	return false, -1
}

func testing() {

	//gameState := newGameState()
	//fmt.Println(gameState.parser("buy firearms 3"))

	myGal := buildGalaxy(1)

	fmt.Println("System 7 is:   " + myGal[7].name)
	fmt.Println("System 7 Desc: " + myGal[7].description)
}

func newElite() elite {
	/* 6502 Elite fires up at Lave with fluctuation=00
	   and these prices tally with the NES ones.
	   However, the availabilities reside in the saved game data.
	   Availabilities are calculated (and fluctuation randomised)
	   on hyperspacing
	   I have checked with this code for Zaonce with fluctaution &AB
	   against the SuperVision 6502 code and both prices and availabilities tally.
	*/

	//testing()
	var elite elite

	elite.state = newGameState()

	elite.state.useNativeRand = true
	elite.state.mySRand(12345) /* Ensure repeatability */

	fmt.Printf("\nWelcome to Text Elite 1.4.\n")

	for i := 0; i <= lastTrade; i++ {

	}

	elite.state.galaxyNum = 1
	elite.state.galaxy = buildGalaxy(elite.state.galaxyNum)

	elite.state.currentPlanet = numForLave /* Don't use jump */

	elite.state.localMarket = newMarket(0x00, elite.state.galaxy[numForLave])
	elite.state.fuel = uint(elite.state.maxFuel)

	elite.state.parser("hold 20") /* Small cargo bay */
	elite.state.parser("cash +100")
	elite.state.parser("help")

	fmt.Printf("\n\nCash :%d> ", elite.state.cash/10)

	return elite
}

func (elite *elite) command(cmd string) {
	elite.state.parser(cmd)

	fmt.Printf("\n\nCash :%d> ", elite.state.cash/10)
}

func main() {
	game := newElite()

	for {
		reader := bufio.NewReader(os.Stdin)
		newCmd, _ := reader.ReadString('\n')
		newCmd = strings.TrimSpace(newCmd)
		fmt.Println("COMMAND ENTERED: " + newCmd)
		game.command(newCmd)
	}
}
