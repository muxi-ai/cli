package validate

import (
	"testing"
)

// FuzzIsValidID tests the isValidID function with random inputs
func FuzzIsValidID(f *testing.F) {
	// Seed corpus with known inputs
	f.Add("my-formation")
	f.Add("test-app-123")
	f.Add("")
	f.Add("-invalid")
	f.Add("UPPERCASE")
	f.Add("has_underscore")
	f.Add("has.dot")
	f.Add("has space")
	f.Add("123start")
	f.Add("a")
	f.Add("valid-id")

	f.Fuzz(func(t *testing.T, id string) {
		// Just ensure it doesn't panic
		_ = isValidID(id)
	})
}

// FuzzCollectSecretRefs tests the collectSecretRefs function with random inputs
func FuzzCollectSecretRefs(f *testing.F) {
	// Seed corpus with known inputs
	f.Add("just some text")
	f.Add("api_key: ${{ secrets.OPENAI_KEY }}")
	f.Add("key1: ${{ secrets.KEY_A }}\nkey2: ${{ secrets.KEY_B }}")
	f.Add("${{ secrets.SAME }} and ${{ secrets.SAME }}")
	f.Add("${{secrets.NO_SPACE}}")
	f.Add("${{  secrets.WITH_SPACE  }}")
	f.Add("malformed ${{ secrets. }}")
	f.Add("nested ${{ secrets.${{ secrets.INNER }} }}")
	f.Add("")

	f.Fuzz(func(t *testing.T, content string) {
		// Just ensure it doesn't panic and returns a valid slice
		result := collectSecretRefs(content)
		if result == nil {
			t.Error("collectSecretRefs should not return nil")
		}
	})
}
