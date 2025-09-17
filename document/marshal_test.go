package document_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/campfhir/wsv/document"
)

type Capital string

func (c *Capital) MarshalWSV(format string) (*string, error) {
	if string(*c) == "Tokyo" && format == "emoji" {
		v := "🗼"
		return &v, nil
	}
	v := string(*c)
	return &v, nil
}

type CountryInfo struct {
	Country               string     `wsv:"Country"`
	Capital               Capital    `wsv:"Capital,format:'emoji'"`
	ExampleLeader         *string    `wsv:"Example Leader"`
	GDP                   float32    `wsv:"GPD,format:'%.1f'"`
	Democracy             bool       `wsv:"Has democracy,format:'true|false'"`
	LastUpdated           *time.Time `wsv:"Last Updated,format:'Jan 02, 2006'"`
	ExamplePoliticalParty *string    `wsv:"Example Party"`
	Emoji                 string     `wsv:"Emoji of Flag"`
	InterestingFact       string     `wsv:"Interesting Facts"`
	Comments              string     `wsv:",comment"`
}

func TestMarshal(t *testing.T) {
	type Example struct {
		A int           `wsv:"A,format:'%#x'"`
		B int16         `wsv:"int16,format:%d"`
		C *int32        `wsv:"int32,format:%x"`
		D int8          `wsv:"eight"`
		T time.Duration `wsv:"Duration"`
		E *float32      `wsv:"-,"`                 // with a comma the field with - does try to marshal
		K int           `wsv:",format:%d,comment"` // formats using %d but adds to comment
		N CountryInfo   `wsv:"-"`                  // this is ignored because there is no comma
	}
	c := int32(2243)
	e := float32(5020.24)
	tt, _ := time.ParseDuration("300ms")
	d, err := document.Marshal([]Example{{
		A: 42,
		B: int16(21),
		C: &c,
		D: 100,
		T: tt,
		K: 1000,
		E: &e,
	}})
	if err != nil {
		t.Fatal(err)
	}
	exp_lines := []string{
		"A     int16  int32  eight  Duration  \"-\"",
		"0x2a  21     8c3    100    300ms     5020.24  #1000",
		``,
	}
	lines := strings.Split(string(d), "\n")
	if len(lines) != len(exp_lines) {
		t.Error("expected", len(exp_lines), "lines but got", len(lines), "instead")
		if len(lines) > len(exp_lines) {
			t.Error("extra lines", lines[len(exp_lines)-1:], exp_lines[len(exp_lines)-1:])
		}
		return
	}
	for i, ln := range lines {
		ex := exp_lines[i]
		if ex != ln {
			t.Error("the line", i+1, "does not have the expected value\n", ex, "!=\n", ln)
		}
	}
}

func TestMarsalSimpleSlice(t *testing.T) {
	example_date, _ := time.Parse(time.DateOnly, "2025-08-16")
	// Leader + Party strings
	macron := "Emmanuel Macron"
	renaissance := "Renaissance"
	scholz := "Olaf Scholz"
	spd := "Social Democratic Party"
	biden := "Joe Biden"
	democrats := "Democratic Party"
	sunak := "Rishi Sunak"
	conservatives := "Conservative Party"
	trudeau := "Justin Trudeau"
	libs := "Liberal Party"
	modi := "Narendra Modi"
	bjp := "Bharatiya Janata Party"
	kishida := "Fumio Kishida"
	ldp := "Liberal Democratic Party"
	lula := "Luiz Inácio Lula da Silva"
	pt := "Workers' Party"
	xi := "Xi Jinping"
	ccp := "Communist Party of China"
	putin := "Vladimir Putin"
	unitedRussia := "United Russia"
	ramaphosa := "Cyril Ramaphosa"
	anc := "African National Congress"
	erdogan := "Recep Tayyip Erdoğan"
	akp := "Justice and Development Party"
	albanese := "Anthony Albanese"
	labor := "Labor Party"
	andersson := "Magdalena Andersson"
	sap := "Social Democratic Party"
	ardern := "Chris Hipkins"
	nzlabour := "Labour Party"
	borch := "Gabriel Boric"
	frente := "Social Convergence"
	amlo := "Andrés Manuel López Obrador"
	morena := "MORENA"
	duda := "Andrzej Duda"
	pis := "Law and Justice"
	zelensky := "Volodymyr Zelenskyy"
	servant := "Servant of the People"
	orbán := "Viktor Orbán"
	fidesz := "Fidesz"
	kenyatta := "William Ruto"
	uda := "United Democratic Alliance"
	salman := "Salman bin Abdulaziz Al Saud"
	saud := "House of Saud"
	countries := []CountryInfo{
		{
			Country:               "France",
			Capital:               "Paris",
			ExampleLeader:         &macron,
			GDP:                   2930.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &renaissance,
			Emoji:                 "🇫🇷",
			InterestingFact:       "The Eiffel Tower was meant to be temporary.",
		},
		{
			Country:               "Germany",
			Capital:               "Berlin",
			ExampleLeader:         &scholz,
			GDP:                   4200.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &spd,
			Emoji:                 "🇩🇪",
			InterestingFact:       "Oktoberfest is the world’s largest beer festival.",
		},
		{
			Country:               "United States",
			Capital:               "Washington D.C.",
			ExampleLeader:         &biden,
			GDP:                   25700.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &democrats,
			Emoji:                 "🇺🇸",
			InterestingFact:       "Alaska has the longest coastline of any U.S. state.",
		},
		{
			Country:               "United Kingdom",
			Capital:               "London",
			ExampleLeader:         &sunak,
			GDP:                   3500.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &conservatives,
			Emoji:                 "🇬🇧",
			InterestingFact:       "Big Ben is the nickname for the bell, not the tower.",
		},
		{
			Country:               "Canada",
			Capital:               "Ottawa",
			ExampleLeader:         &trudeau,
			GDP:                   2300.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &libs,
			Emoji:                 "🇨🇦",
			InterestingFact:       "Canada has the most lakes in the world.",
		},
		{
			Country:               "India",
			Capital:               "New Delhi",
			ExampleLeader:         &modi,
			GDP:                   3760.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &bjp,
			Emoji:                 "🇮🇳",
			InterestingFact:       "The Kumbh Mela gathering is visible from space.",
		},
		{
			Country:               "Japan",
			Capital:               "Tokyo",
			ExampleLeader:         &kishida,
			GDP:                   4200.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &ldp,
			Emoji:                 "🇯🇵",
			InterestingFact:       "There are more pets than children in Japan.",
		},
		{
			Country:               "Brazil",
			Capital:               "Brasília",
			ExampleLeader:         &lula,
			GDP:                   2570.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &pt,
			Emoji:                 "🇧🇷",
			InterestingFact:       "The Amazon rainforest produces 20% of the world’s oxygen.",
		},
		{
			Country:               "China",
			Capital:               "Beijing",
			ExampleLeader:         &xi,
			GDP:                   17700.0,
			Democracy:             false,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &ccp,
			Emoji:                 "🇨🇳",
			InterestingFact:       "China has only one time zone.",
		},
		{
			Country:               "Russia",
			Capital:               "Moscow",
			ExampleLeader:         &putin,
			GDP:                   1740.0,
			Democracy:             false,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &unitedRussia,
			Emoji:                 "🇷🇺",
			InterestingFact:       "Russia spans 11 time zones.",
		},
		{
			Country:               "South Africa",
			Capital:               "Pretoria",
			ExampleLeader:         &ramaphosa,
			GDP:                   399.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &anc,
			Emoji:                 "🇿🇦",
			InterestingFact:       "It has three capital cities.",
		},
		{
			Country:               "Turkey",
			Capital:               "Ankara",
			ExampleLeader:         &erdogan,
			GDP:                   905.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &akp,
			Emoji:                 "🇹🇷",
			InterestingFact:       "Istanbul is the only city on two continents.",
		},
		{
			Country:               "Australia",
			Capital:               "Canberra",
			ExampleLeader:         &albanese,
			GDP:                   1700.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &labor,
			Emoji:                 "🇦🇺",
			InterestingFact:       "Australia has more kangaroos than people.",
		},

		{
			Country:               "Sweden",
			Capital:               "Stockholm",
			ExampleLeader:         &andersson,
			GDP:                   635.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &sap,
			Emoji:                 "🇸🇪",
			InterestingFact:       "Sweden has 95,700 lakes.",
		},
		{
			Country:               "New Zealand",
			Capital:               "Wellington",
			ExampleLeader:         &ardern,
			GDP:                   247.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &nzlabour,
			Emoji:                 "🇳🇿",
			InterestingFact:       "There are more sheep than people.",
		},
		{
			Country:               "Chile",
			Capital:               "Santiago",
			ExampleLeader:         &borch,
			GDP:                   410.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &frente,
			Emoji:                 "🇨🇱",
			InterestingFact:       "Chile is the world’s longest country north-to-south.",
		},
		{
			Country:               "Mexico",
			Capital:               "Mexico City",
			ExampleLeader:         &amlo,
			GDP:                   1670.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &morena,
			Emoji:                 "🇲🇽",
			InterestingFact:       "Mexico introduced chocolate to the world.",
		},
		{
			Country:               "Poland",
			Capital:               "Warsaw",
			ExampleLeader:         &duda,
			GDP:                   842.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &pis,
			Emoji:                 "🇵🇱",
			InterestingFact:       "Poland has one of the oldest universities (1364).",
		},
		{
			Country:               "Ukraine",
			Capital:               "Kyiv",
			ExampleLeader:         &zelensky,
			GDP:                   200.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &servant,
			Emoji:                 "🇺🇦",
			InterestingFact:       "Ukraine is the largest country in Europe fully within Europe.",
		},
		{
			Country:               "Hungary",
			Capital:               "Budapest",
			ExampleLeader:         &orbán,
			GDP:                   210.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &fidesz,
			Emoji:                 "🇭🇺",
			InterestingFact:       "Budapest has more thermal springs than any other capital.",
		},
		{
			Country:               "Kenya",
			Capital:               "Nairobi",
			ExampleLeader:         &kenyatta,
			GDP:                   110.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &uda,
			Emoji:                 "🇰🇪",
			InterestingFact:       "Kenya is the birthplace of coffee cultivation.",
		},
		{
			Country:               "Saudi Arabia",
			Capital:               "Riyadh",
			ExampleLeader:         &salman,
			GDP:                   1040.0,
			Democracy:             false,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: &saud,
			Emoji:                 "🇸🇦",
			InterestingFact:       "Saudi Arabia has no rivers.",
		},
		{
			Country:               "Norway",
			Capital:               "Oslo",
			ExampleLeader:         nil,
			GDP:                   580.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: nil,
			Emoji:                 "🇳🇴",
			InterestingFact:       "Norway introduced salmon sushi to Japan.",
			Comments:              "No party data",
		},
		{
			Country:               "Italy",
			Capital:               "Rome",
			ExampleLeader:         nil,
			GDP:                   2100.0,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: nil,
			Emoji:                 "🇮🇹",
			InterestingFact:       "Italy has more UNESCO World Heritage Sites than any other country.",
			Comments:              "PM changes frequently",
		},
		{
			Country:               "Spain",
			Capital:               "Madrid",
			ExampleLeader:         nil,
			GDP:                   1850.57,
			Democracy:             true,
			LastUpdated:           &example_date,
			ExamplePoliticalParty: nil,
			Emoji:                 "🇪🇸",
			InterestingFact:       "Spain has the world’s oldest restaurant (opened in 1725).",
			Comments:              "Monarch coexists with government",
		},
	}

	d, err := document.Marshal(countries)
	if err != nil {
		t.Fatal(err)
		return
	}
	exp_lines := []string{
		`Country           Capital            "Example Leader"                GPD      "Has democracy"  "Last Updated"  "Example Party"                  "Emoji of Flag"  "Interesting Facts"`,
		`France            Paris              "Emmanuel Macron"               2930.0   true             "Aug 16, 2025"  Renaissance                      🇫🇷               "The Eiffel Tower was meant to be temporary."`,
		`Germany           Berlin             "Olaf Scholz"                   4200.0   true             "Aug 16, 2025"  "Social Democratic Party"        🇩🇪               "Oktoberfest is the world’s largest beer festival."`,
		`"United States"   "Washington D.C."  "Joe Biden"                     25700.0  true             "Aug 16, 2025"  "Democratic Party"               🇺🇸               "Alaska has the longest coastline of any U.S. state."`,
		`"United Kingdom"  London             "Rishi Sunak"                   3500.0   true             "Aug 16, 2025"  "Conservative Party"             🇬🇧               "Big Ben is the nickname for the bell, not the tower."`,
		`Canada            Ottawa             "Justin Trudeau"                2300.0   true             "Aug 16, 2025"  "Liberal Party"                  🇨🇦               "Canada has the most lakes in the world."`,
		`India             "New Delhi"        "Narendra Modi"                 3760.0   true             "Aug 16, 2025"  "Bharatiya Janata Party"         🇮🇳               "The Kumbh Mela gathering is visible from space."`,
		`Japan             🗼                  "Fumio Kishida"                 4200.0   true             "Aug 16, 2025"  "Liberal Democratic Party"       🇯🇵               "There are more pets than children in Japan."`,
		`Brazil            Brasília           "Luiz Inácio Lula da Silva"     2570.0   true             "Aug 16, 2025"  "Workers' Party"                 🇧🇷               "The Amazon rainforest produces 20% of the world’s oxygen."`,
		`China             Beijing            "Xi Jinping"                    17700.0  false            "Aug 16, 2025"  "Communist Party of China"       🇨🇳               "China has only one time zone."`,
		`Russia            Moscow             "Vladimir Putin"                1740.0   false            "Aug 16, 2025"  "United Russia"                  🇷🇺               "Russia spans 11 time zones."`,
		`"South Africa"    Pretoria           "Cyril Ramaphosa"               399.0    true             "Aug 16, 2025"  "African National Congress"      🇿🇦               "It has three capital cities."`,
		`Turkey            Ankara             "Recep Tayyip Erdoğan"          905.0    true             "Aug 16, 2025"  "Justice and Development Party"  🇹🇷               "Istanbul is the only city on two continents."`,
		`Australia         Canberra           "Anthony Albanese"              1700.0   true             "Aug 16, 2025"  "Labor Party"                    🇦🇺               "Australia has more kangaroos than people."`,
		`Sweden            Stockholm          "Magdalena Andersson"           635.0    true             "Aug 16, 2025"  "Social Democratic Party"        🇸🇪               "Sweden has 95,700 lakes."`,
		`"New Zealand"     Wellington         "Chris Hipkins"                 247.0    true             "Aug 16, 2025"  "Labour Party"                   🇳🇿               "There are more sheep than people."`,
		`Chile             Santiago           "Gabriel Boric"                 410.0    true             "Aug 16, 2025"  "Social Convergence"             🇨🇱               "Chile is the world’s longest country north-to-south."`,
		`Mexico            "Mexico City"      "Andrés Manuel López Obrador"   1670.0   true             "Aug 16, 2025"  MORENA                           🇲🇽               "Mexico introduced chocolate to the world."`,
		`Poland            Warsaw             "Andrzej Duda"                  842.0    true             "Aug 16, 2025"  "Law and Justice"                🇵🇱               "Poland has one of the oldest universities (1364)."`,
		`Ukraine           Kyiv               "Volodymyr Zelenskyy"           200.0    true             "Aug 16, 2025"  "Servant of the People"          🇺🇦               "Ukraine is the largest country in Europe fully within Europe."`,
		`Hungary           Budapest           "Viktor Orbán"                  210.0    true             "Aug 16, 2025"  Fidesz                           🇭🇺               "Budapest has more thermal springs than any other capital."`,
		`Kenya             Nairobi            "William Ruto"                  110.0    true             "Aug 16, 2025"  "United Democratic Alliance"     🇰🇪               "Kenya is the birthplace of coffee cultivation."`,
		`"Saudi Arabia"    Riyadh             "Salman bin Abdulaziz Al Saud"  1040.0   false            "Aug 16, 2025"  "House of Saud"                  🇸🇦               "Saudi Arabia has no rivers."`,
		`Norway            Oslo               -                               580.0    true             "Aug 16, 2025"  -                                🇳🇴               "Norway introduced salmon sushi to Japan."  #No party data`,
		`Italy             Rome               -                               2100.0   true             "Aug 16, 2025"  -                                🇮🇹               "Italy has more UNESCO World Heritage Sites than any other country."  #PM changes frequently`,
		`Spain             Madrid             -                               1850.6   true             "Aug 16, 2025"  -                                🇪🇸               "Spain has the world’s oldest restaurant (opened in 1725)."  #Monarch coexists with government`,
		``, // the trailing new line
	}
	lines := strings.Split(string(d), "\n")
	if len(lines) != len(exp_lines) {
		t.Error("expected", len(exp_lines), "lines but got", len(lines), "instead")
		if len(lines) > len(exp_lines) {
			t.Error("extra lines", lines[len(exp_lines)-1:], exp_lines[len(exp_lines)-1:])
		}
		return
	}
	for i, ln := range lines {
		ex := exp_lines[i]
		if ex != ln {
			t.Error("the line", i+1, "does not have the expected value\n", ex, "!=\n", ln)
		}
	}
}

func TestMarshalMultipleComments(t *testing.T) {
	type Person struct {
		FirstName string `wsv:"First Name"`
		LastName  string `wsv:"Last Name"`
		Fact1     string `wsv:",comment"` // field name does not matter with the comment attribute
		Fact2     string `wsv:",comment"`
	}
	d, err := document.Marshal([]Person{
		{
			FirstName: "Scott",
			LastName:  "Eremia-Roden",
			Fact1:     "Wrote this program",
			Fact2:     "Wrote a test once",
		},
		{
			FirstName: "John",
			LastName:  "Doe",
			Fact2:     "Did a thing",
			Fact1:     "Opened a thing",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	exp_lines := []string{
		`"First Name"  "Last Name"`,
		`Scott         "Eremia-Roden"  #Wrote this program Wrote a test once`,
		`John          Doe  #Opened a thing Did a thing`,
		``,
	}

	lines := strings.Split(string(d), "\n")
	if len(lines) != len(exp_lines) {
		t.Error("expected", len(exp_lines), "lines but got", len(lines), "instead")
		if len(lines) > len(exp_lines) {
			t.Error("extra lines", lines[len(exp_lines)-1:], exp_lines[len(exp_lines)-1:])
		}
		return
	}
	for i, ln := range lines {
		ex := exp_lines[i]
		if ex != ln {
			t.Error("the line", i+1, "does not have the expected value\n", ex, "!=\n", ln)
		}
	}
}

func TestMarshalString(t *testing.T) {
	type Example struct {
		Str string `wsv:"String Value"`
	}
	b, err := document.Marshal([]Example{{"a value"}})
	if err != nil {
		t.Fatal(err)
	}
	exp := []byte(strings.Join([]string{
		"String Value",
		"a value",
		"\n",
	}, "\n"))
	if bytes.Equal(b, exp) {
		t.Error("not the same")
	}
}

func TestMarshalStringPtr(t *testing.T) {
	type Example struct {
		Str *string `wsv:"String Value"`
	}
	ptr := "a value"
	b, err := document.Marshal([]Example{{Str: &ptr}})
	if err != nil {
		t.Fatal(err)
	}
	exp := []byte(strings.Join([]string{
		"String Value",
		"a value",
		"\n",
	}, "\n"))
	if bytes.Equal(b, exp) {
		t.Error("not the same")
	}
}

func TestMarshalInt(t *testing.T) {
	type Example struct {
		Number string `wsv:"Num"`
	}
	b, err := document.Marshal([]Example{{"6"}})
	if err != nil {
		t.Fatal(err)
	}
	exp := []byte(strings.Join([]string{
		"Num",
		"6",
		"\n",
	}, "\n"))
	if bytes.Equal(b, exp) {
		t.Error("not the same")
	}
}

func TestMarshalIntPtr(t *testing.T) {
	type Example struct {
		Number *int `wsv:"Num"`
	}
	ptr := 6
	b, err := document.Marshal([]Example{{Number: &ptr}})
	if err != nil {
		t.Fatal(err)
	}
	exp := []byte(strings.Join([]string{
		"Num",
		"6",
		"\n",
	}, "\n"))
	if bytes.Equal(b, exp) {
		t.Error("not the same")
	}
}
