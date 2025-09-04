package internal_test

import (
	"testing"

	"github.com/campfhir/wsv/internal"
)

func TestSerializeText(t *testing.T) {
	rec1 := internal.Field{
		Value: "Japan is a volcanic archipelago with over 100 active volcanoes.\nThe currency is the yen and the symbol is ¥.",
	}
	exp1 := `"Japan is a volcanic archipelago with over 100 active volcanoes."/"The currency is the yen and the symbol is ¥."`
	out1 := rec1.SerializeText()
	if out1 != exp1 {
		t.Errorf("expect\n%s\nbut got\n%s\ninstead", exp1, out1)
	}
	cal1 := rec1.CalculateFieldLength()
	if cal1 != 112 {
		t.Error(cal1)
	}

	rec2 := internal.Field{
		Value: "Would you've guessed that vodka or gin tops the list? For years, Jinro Soju has been the world's best-selling alcohol! It might not be surprising, given that with 11.2 shots on average, Koreans are also the world's biggest consumer of hard liquor. Haven't been able to try it yet? Time to visit Korea!",
	}
	exp2 := `"Would you've guessed that vodka or gin tops the list? For years, Jinro Soju has been the world's best-selling alcohol! It might not be surprising, given that with 11.2 shots on average, Koreans are also the world's biggest consumer of hard liquor. Haven't been able to try it yet? Time to visit Korea!"`
	out2 := rec2.SerializeText()
	if out2 != exp2 {
		t.Errorf("expect\n%s\nbut got\n%s\ninstead", exp2, out2)
	}
}
