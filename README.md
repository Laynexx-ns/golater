<img src="./img/golater_preview.png" alt="logo">

## Installation

simply run this command

```bash
go install github.com/Laynexx-ns/golater@latest
```

## Usage

After installing project, you can run it via `golater` command in terminal. Default path for all
configurations is `$HOME+.gonfig/golater/golater.json`, there you can edit your templates. <br/>

`.config/golater/golater.json` creates automatically if not exist. JSON scheme described in this repository (`json.scheme.json`), also you can find it in source files structures

### Keywords and other

#### filename/dir

`$n` - golater replaces it into it's number index <br/>
for example: if I say golater that I want to spawn 3 repeated parts (where repeated part is just one file - `file-$n.txt`) -> <br/>

```
|- file-1.txt
|- file-2.txt
|- file-3.txt
```

`/path-a/path-b/path-c/file.txt` - golater converts it into right path - 3 directories and one file. <br/>
This rule works in every "folder" and "filename" field

### Creating your first template

here's example template:

```json
{
  "path": "~/.config/golater/golater.json",
  "templates": [
    {
      "name": "C++ + Python tests template",
      "description": "my custom template for cpp algorithms solving",
      "lang": "cpp,py",
      "endline_format": "\n",
      "repeated": {
        "folder": "task-$n",
        "files": [
          {
            "filename": "solution-$n",
            "ext": "cpp",
            "data": [
              "#include <iostream>",
              "",
              "int main(){",
              "  std::ios::sync_with_stdio(false);",
              "  std::cin.tie(nullptr);",
              "  std::cout << 3;",
              "  std::cout << \"meow\";",
              "}"
            ]
          },
          {
            "filename": "task-$n",
            "ext": "md",
            "data": []
          },
          {
            "filename": "python-tests/test-$n",
            "ext": "py",
            "data": ["print(3)"]
          },
          {
            "filename": "tests/test-$n",
            "ext": "txt",
            "data": []
          }
        ]
      },
      "root": [
        {
          "filename": "",
          "ext": "clang-format",
          "data": [
            "BasedOnStyle: Google",
            "IndentAccessModifiers: false",
            "AccessModifierOffset: -1"
          ]
        }
      ]
    }
  ]
}
```
