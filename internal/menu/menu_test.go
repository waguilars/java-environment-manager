package menu

import (
	"testing"
)

func TestNewModel(t *testing.T) {
	m := NewModel()

	if len(m.items) != 7 {
		t.Errorf("Expected 7 menu items, got %d", len(m.items))
	}

	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", m.cursor)
	}

	expectedActions := []string{"setup", "scan", "list", "current", "use", "install", "exit"}
	for i, item := range m.items {
		if item.Action != expectedActions[i] {
			t.Errorf("Item %d: expected action %s, got %s", i, expectedActions[i], item.Action)
		}
	}
}

func TestModelUpdateNavigation(t *testing.T) {
	m := NewModel()

	// Test moving down
	_, _ = m.Update(nil)
	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0 after Init, got %d", m.cursor)
	}

	// Test moving down
	_, _ = m.Update(nil)
	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0 after no key, got %d", m.cursor)
	}
}

func TestUseModel(t *testing.T) {
	jdkList := []JDKInfo{
		{Name: "Temurin", Version: "17.0.1", Provider: "temurin", Managed: true, Path: "/usr/lib/jvm/temurin-17"},
		{Name: "Temurin", Version: "21.0.1", Provider: "temurin", Managed: true, Path: "/usr/lib/jvm/temurin-21"},
	}

	m := NewUseModel(jdkList, "Select a JDK")

	if len(m.jdkList) != 2 {
		t.Errorf("Expected 2 JDKs, got %d", len(m.jdkList))
	}

	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", m.cursor)
	}
}

func TestInstallModel(t *testing.T) {
	versions := []InstallableVersion{
		{Version: "17.0.1", Major: 17, LTS: true, Available: true},
		{Version: "21.0.1", Major: 21, LTS: false, Available: true},
		{Version: "22.0.1", Major: 22, LTS: false, Available: true},
	}

	m := NewInstallModel(versions, "Select a version to install")

	if len(m.versions) != 3 {
		t.Errorf("Expected 3 versions, got %d", len(m.versions))
	}

	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", m.cursor)
	}
}

func TestInstallModelLTSToggle(t *testing.T) {
	versions := []InstallableVersion{
		{Version: "17.0.1", Major: 17, LTS: true, Available: true},
		{Version: "21.0.1", Major: 21, LTS: false, Available: true},
		{Version: "22.0.1", Major: 22, LTS: false, Available: true},
	}

	m := NewInstallModel(versions, "Select a version to install")

	if m.filterLTS {
		t.Error("Expected filterLTS to be false initially")
	}

	m.filterLTS = true

	if !m.filterLTS {
		t.Error("Expected filterLTS to be true after setting")
	}
}

func TestRunUseMenuEmptyList(t *testing.T) {
	_, err := RunUseMenu([]JDKInfo{}, "No JDKs available")

	if err == nil {
		t.Error("Expected error for empty JDK list")
	}
}

func TestRunInstallMenuEmptyList(t *testing.T) {
	_, err := RunInstallMenu([]InstallableVersion{}, "No versions available")

	if err == nil {
		t.Error("Expected error for empty version list")
	}
}

func TestJDKInfoString(t *testing.T) {
	jdk := JDKInfo{
		Name:     "Temurin",
		Version:  "17.0.1",
		Provider: "temurin",
		Managed:  true,
		Path:     "/usr/lib/jvm/temurin-17",
	}

	if jdk.Version != "17.0.1" {
		t.Errorf("Expected version 17.0.1, got %s", jdk.Version)
	}

	if !jdk.Managed {
		t.Error("Expected JDK to be managed")
	}
}

func TestInstallableVersion(t *testing.T) {
	v := InstallableVersion{
		Version: "17.0.1",
		Major:   17,
		LTS:     true,
	}

	if v.Version != "17.0.1" {
		t.Errorf("Expected version 17.0.1, got %s", v.Version)
	}

	if !v.LTS {
		t.Error("Expected version to be LTS")
	}

	if v.Major != 17 {
		t.Errorf("Expected major 17, got %d", v.Major)
	}
}
