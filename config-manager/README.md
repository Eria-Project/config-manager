[![Build Status](https://travis-ci.org/Eria-Project/config-manager.svg?branch=master)](https://travis-ci.org/Eria-Project/config-manager)

# ERIA Project config manager (Work In Progress)

## Init the manager
```
var config struct {
	A string `default:"A"`
	B uint   `default:"1"`
	C bool   `default:"true"`
	D struct {
		D1 string
	}
	E []struct {
		E1 string
	}
	F string `required:"true"`
}

cm, err := configmanager.Init(configFile, &config)
if err != nil {
    os.Exit(1)
}
defer cm.Close()

```

## Load the config from the file
```
if err := cm.Load(); err != nil {
    os.Exit(1)
}
```

## Save the config to the file
```
if err := cm.Save(); err != nil {
    os.Exit(1)
}
```

## Monitor file changes
```
for {
    cm.Next()
    fmt.Println("Update")
}
```

## Monitor a specific value changes
```
w := cm.Watch("D.D1")

for {
    w.Next()
    fmt.Println("Update")
}
```