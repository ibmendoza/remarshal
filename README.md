Convenience wrapper of https://github.com/dbohdan/remarshal which converts his CLI app to library.

API
---

```go
func Convert(input []byte, inputF, outputF string) (string, error)
```

where 

- input - convert your input string to []byte
- inputF - can be TOML, JSON or YAML
- outputF - can be TOML, JSON or YAML

It is up to you if you want to write the converted string to a file which is trivial.

Thanks Danyil Bohdan (https://github.com/dbohdan/remarshal)
