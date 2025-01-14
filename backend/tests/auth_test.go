package main_tests

import (
	"github.com/devs-group/sloth/backend/services"
	"github.com/markbates/goth"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoginWithDifferentProvidersResolvingSameEmail(t *testing.T) {
	dbService := SetupTestEnvironment(t)
	conn := dbService.GetConn()
	defer conn.Close()
	defer dbService.Delete()

	tx := conn.MustBegin()

	userID, err := services.UpsertUserBySocialIDAndMethod("github", &goth.User{
		Provider: "github",
		Email:    "john@doe.com",
		NickName: "johndoe",
		UserID:   "48172313",
	}, tx)
	assert.NoError(t, err)

	assert.Equal(t, 1, userID)

	userID, err = services.UpsertUserBySocialIDAndMethod("google", &goth.User{
		Provider: "google",
		Email:    "john@doe.com",
		NickName: "dohnjoe",
		UserID:   "11111111",
	}, tx)
	assert.NoError(t, err)

	assert.Equal(t, 1, userID)

	err = tx.Commit()
	assert.NoError(t, err)

	var count int
	err = conn.QueryRow("SELECT COUNT(*) FROM auth_methods WHERE user_id = $1", userID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestLoginWithDifferentProvidersResolvingDifferentMail(t *testing.T) {
	dbService := SetupTestEnvironment(t)
	conn := dbService.GetConn()
	defer conn.Close()
	defer dbService.Delete()

	tx := conn.MustBegin()

	userID, err := services.UpsertUserBySocialIDAndMethod("github", &goth.User{
		Provider: "github",
		Email:    "john@doe.com",
		NickName: "johndoe",
		UserID:   "48172313",
	}, tx)
	assert.NoError(t, err)

	assert.Equal(t, 1, userID)

	userID, err = services.UpsertUserBySocialIDAndMethod("google", &goth.User{
		Provider: "google",
		Email:    "peter@jansen.com",
		NickName: "jansen1972",
		UserID:   "9318231",
	}, tx)
	assert.NoError(t, err)

	assert.Equal(t, 2, userID)

	err = tx.Commit()
	assert.NoError(t, err)

	var count int
	err = conn.QueryRow("SELECT COUNT(*) FROM auth_methods WHERE user_id = $1", userID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}
