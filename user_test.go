package main

// func TestReadUsersFileNotExist(t *testing.T) {
// 	err := readFileUsers("./does_not_exist")
// 	assert.NotEqual(t, err, nil, "ReadUsers() read file ./does_not_exist")
// }

// func setupTempUsersTestFile(f string) []User {
// 	users.Range(func(k, v interface{}) bool {
// 		meters.Delete(k)
// 		return true
// 	})

// 	us := []User{
// 		{"abc", "123", "x-xx"},
// 		{"def", "456", "x-xx"},
// 		{"ghi", "789", "x-xx"},
// 	}
// 	for _, u := range us {
// 		users.Store(u.Username, u)
// 	}

// 	writeFileUsers(fnameUsers)

// 	return us
// }
// func TestReadUsers(t *testing.T) {
// 	us := setupTempUsersTestFile("./tmpuser1")

// 	readFileUsers("./tmpuser1")
// 	u0, _ := users.Load(us[0].Username)
// 	assert.Equal(t, u0.(User), us[0], "users[0] != us[0]")
// 	u1, _ := users.Load(us[1].Username)
// 	assert.Equal(t, u1.(User), us[1], "users[1] != us[1]")

// 	os.Remove("./tmpuser1")
// }
