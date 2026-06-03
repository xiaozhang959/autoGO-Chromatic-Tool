package main

import "testing"

func TestAndroidCapDexAppInfoArgsUsesClasspathAppProcess(t *testing.T) {
	got := androidCapDexAppInfoArgs()
	want := []string{
		"shell",
		"CLASSPATH=" + androidCapDexDevicePath,
		"app_process",
		"/",
		androidCapDexAppInfoClass,
	}
	assertStringSliceEqual(t, got, want)
}

func TestParseAndroidAppInfoLines(t *testing.T) {
	output := `
ignored warning
{"name":"微信","packageName":"com.tencent.mm","activityName":"com.tencent.mm.ui.LauncherUI"}
{"name":"","packageName":"com.example.no_label","activityName":""}
`

	apps, err := parseAndroidAppInfoLines(output)
	if err != nil {
		t.Fatalf("parseAndroidAppInfoLines failed: %v", err)
	}
	if len(apps) != 2 {
		t.Fatalf("app count mismatch: %d", len(apps))
	}
	if apps[0].Name != "微信" || apps[0].PackageName != "com.tencent.mm" || apps[0].ActivityName != "com.tencent.mm.ui.LauncherUI" {
		t.Fatalf("first app mismatch: %+v", apps[0])
	}
	if apps[1].Name != "com.example.no_label" {
		t.Fatalf("empty app name should fallback to package name: %+v", apps[1])
	}
}

func TestParseAndroidAppInfoLinesRejectsMissingPackage(t *testing.T) {
	_, err := parseAndroidAppInfoLines(`{"name":"坏数据","packageName":"","activityName":""}`)
	if err == nil {
		t.Fatal("expected missing package error")
	}
}

func TestFilterAndroidAppsMatchesNamePackageAndActivity(t *testing.T) {
	apps := []AndroidAppInfo{
		{Name: "微信", PackageName: "com.tencent.mm", ActivityName: "com.tencent.mm.ui.LauncherUI"},
		{Name: "Settings", PackageName: "com.android.settings", ActivityName: "com.android.settings.Settings"},
	}

	if got := filterAndroidApps(apps, "微信"); len(got) != 1 || got[0].PackageName != "com.tencent.mm" {
		t.Fatalf("name filter mismatch: %+v", got)
	}
	if got := filterAndroidApps(apps, "ANDROID.SETTINGS"); len(got) != 1 || got[0].PackageName != "com.android.settings" {
		t.Fatalf("package filter mismatch: %+v", got)
	}
	if got := filterAndroidApps(apps, "launcherui"); len(got) != 1 || got[0].PackageName != "com.tencent.mm" {
		t.Fatalf("activity filter mismatch: %+v", got)
	}
}
