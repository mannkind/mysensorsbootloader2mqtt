package main

type nodeSettings struct {
	Type    uint16
	Version uint16
}

type nodeSettingsMap map[string]nodeSettings
