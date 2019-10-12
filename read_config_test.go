package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestReadYamlConfig(t *testing.T) {
	email := "email@gmail.com"
	password := "password"
	player := "mplayer"
	logLevel := "ERROR"
	config := fmt.Sprintf(`
email: %s
password: %s
add: false
log_level: %s
sound: true
player: %s
`,
		email,
		password,
		logLevel,
		player,
	)
	args := &lingualeoArgs{}
	decoder := newConfigFile("test.yaml")
	require.Equal(t, yamlType, decoder.getType())
	err := decoder.decode([]byte(config), args)
	require.NoError(t, err, config)
	assert.Equal(t, email, args.Email)
	assert.Equal(t, password, args.Password)
	assert.Equal(t, player, args.Player)
	assert.Equal(t, logLevel, args.LogLevel)
	assert.True(t, args.Sound)
	assert.False(t, args.Add)
}

func TestReadJsonConfig(t *testing.T) {
	email := "email@gmail.com"
	password := "password"
	player := "mplayer"
	logLevel := "ERROR"
	config := fmt.Sprintf(`{
"email": "%s",
"password": "%s",
"add": false,
"log_level": "%s",
"sound": true,
"player": "%s"
}`,
		email,
		password,
		logLevel,
		player,
	)
	args := &lingualeoArgs{}
	decoder := newConfigFile("test.json")
	require.Equal(t, jsonType, decoder.getType())
	err := decoder.decode([]byte(config), args)
	require.NoError(t, err, config)
	assert.Equal(t, email, args.Email)
	assert.Equal(t, password, args.Password)
	assert.Equal(t, player, args.Player)
	assert.Equal(t, logLevel, args.LogLevel)
	assert.True(t, args.Sound)
	assert.False(t, args.Add)
}

func TestReadTomlConfig(t *testing.T) {
	email := "email@gmail.com"
	password := "password"
	player := "mplayer"
	logLevel := "ERROR"
	config := fmt.Sprintf(`
email = "%s"
password = "%s"
add = false
log_level = "%s"
sound = true
player = "%s"
`,
		email,
		password,
		logLevel,
		player,
	)
	args := &lingualeoArgs{}
	decoder := newConfigFile("test.ini")
	require.Equal(t, tomlType, decoder.getType())
	err := decoder.decode([]byte(config), args)
	require.NoError(t, err, config)
	assert.Equal(t, email, args.Email)
	assert.Equal(t, password, args.Password)
	assert.Equal(t, player, args.Player)
	assert.Equal(t, logLevel, args.LogLevel)
	assert.True(t, args.Sound)
	assert.False(t, args.Add)
}

func TestReadManyConfig(t *testing.T) {
	email1 := "email1@gmail.com"
	email2 := "email2@gmail.com"
	password := "password"
	player := "mplayer"
	logLevel := "ERROR"
	tomlConfig := fmt.Sprintf(`
email = "%s"
password = "%s"
`,
		email1,
		password,
	)
	yamlConfig := fmt.Sprintf(`
email: %s
add: true
log_level: %s
sound: true
player: "%s"
`,
		email2,
		logLevel,
		player,
	)
	args := &lingualeoArgs{}

	decoder := newConfigFile("test.ini")
	require.Equal(t, tomlType, decoder.getType())
	err := decoder.decode([]byte(tomlConfig), args)
	require.NoError(t, err, tomlConfig)

	assert.Equal(t, email1, args.Email)
	assert.Equal(t, password, args.Password)
	assert.Equal(t, "", args.Player)
	assert.Equal(t, "", args.LogLevel)
	assert.False(t, args.Sound)
	assert.False(t, args.Add)

	decoder = newConfigFile("test.yaml")
	require.Equal(t, yamlType, decoder.getType())
	err = decoder.decode([]byte(yamlConfig), args)
	require.NoError(t, err, yamlConfig)

	assert.Equal(t, email2, args.Email)
	assert.Equal(t, password, args.Password)
	assert.Equal(t, player, args.Player)
	assert.Equal(t, logLevel, args.LogLevel)
	assert.True(t, args.Sound)
	assert.True(t, args.Add)
}
