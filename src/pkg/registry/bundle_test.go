package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}

func TestCreateBundleIncludesTriggersAndAuthScripts(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "formation.afs"), "schema: \"1.0.0\"\nid: auth-form\n", 0644)
	writeFile(t, filepath.Join(src, "triggers", "plain.md"), "Event: ${{ data.type }}\n", 0644)
	writeFile(t, filepath.Join(src, "triggers", "stripe.md"), "---\nauth: ./auth/stripe.py\n---\nEvent: ${{ data.type }}\n", 0644)
	writeFile(t, filepath.Join(src, "triggers", "hubspot.md"), "---\nauth: https://auth.example.com/hubspot\n---\nHubSpot\n", 0644)
	writeFile(t, filepath.Join(src, "triggers", "escape.md"), "---\nauth: ../outside.py\n---\nNope\n", 0644)
	writeFile(t, filepath.Join(src, "auth", "stripe.py"), "#!/usr/bin/env python3\n", 0755)
	writeFile(t, filepath.Join(src, "auth", "unreferenced.py"), "#!/usr/bin/env python3\n", 0755)

	info, err := CreateBundle(src)
	if err != nil {
		t.Fatal(err)
	}
	defer CleanupBundle(info.Path)

	dest := t.TempDir()
	if _, err := ExtractBundle(info.Path, dest); err != nil {
		t.Fatal(err)
	}

	for _, rel := range []string{"formation.afs", "triggers/plain.md", "triggers/stripe.md", "triggers/hubspot.md", "triggers/escape.md", "auth/stripe.py"} {
		if _, err := os.Stat(filepath.Join(dest, rel)); err != nil {
			t.Errorf("expected %s in bundle: %v", rel, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dest, "auth", "unreferenced.py")); !os.IsNotExist(err) {
		t.Errorf("auth/unreferenced.py is not named by any trigger and should not be bundled")
	}

	stat, err := os.Stat(filepath.Join(dest, "auth", "stripe.py"))
	if err != nil {
		t.Fatal(err)
	}
	if stat.Mode().Perm()&0100 == 0 {
		t.Errorf("auth/stripe.py lost its executable bit: %v", stat.Mode())
	}
}

func TestFrontmatterAuth(t *testing.T) {
	cases := map[string]string{
		"---\nauth: ./auth/a.py\n---\nbody":            "./auth/a.py",
		"---\r\nauth: ./auth/a.py\r\n---\r\nbody":      "./auth/a.py",
		"---\nname: x\nauth: https://a.example\n---\n": "https://a.example",
		"---\nname: x\n---\nauth: ./not-frontmatter":   "",
		"no frontmatter":              "",
		"---\nauth: [unclosed\n---\n": "",
	}
	for in, want := range cases {
		if got := frontmatterAuth(in); got != want {
			t.Errorf("frontmatterAuth(%q) = %q, want %q", in, got, want)
		}
	}
}
