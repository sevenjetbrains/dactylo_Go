package text

import "testing"

func TestNFCComposesDeadKeySequences(t *testing.T) {
	decomposed := "être" // e + accent circonflexe combinant
	if got := NFC(decomposed); got != "être" {
		t.Errorf("NFC(%q) = %q", decomposed, got)
	}
	if n := len([]rune(NFC(decomposed))); n != 4 {
		t.Errorf("longueur en runes = %d, attendu 4", n)
	}
}

func TestFoldTypography(t *testing.T) {
	cases := map[rune]rune{
		'’': '\'', '‘': '\'', '«': '"', '»': '"', '“': '"', '”': '"',
		'–': '-', '—': '-', 'a': 'a', 'é': 'é', '\'': '\'',
	}
	for in, want := range cases {
		if got := FoldTypography(in); got != want {
			t.Errorf("FoldTypography(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsSpaceLike(t *testing.T) {
	for _, r := range []rune{' ', ' ', ' ', '\t'} {
		if !IsSpaceLike(r) {
			t.Errorf("%U devrait valoir une espace", r)
		}
	}
	for _, r := range []rune{'a', '\n', '-'} {
		if IsSpaceLike(r) {
			t.Errorf("%U ne devrait pas valoir une espace", r)
		}
	}
}

func TestCleanForPractice(t *testing.T) {
	in := "  Premier\tparagraphe.  \r\n\r\n\r\n\r\nSecond être. \n"
	want := "Premier paragraphe.\n\nSecond être."
	if got := CleanForPractice(in); got != want {
		t.Errorf("CleanForPractice = %q, want %q", got, want)
	}
}
