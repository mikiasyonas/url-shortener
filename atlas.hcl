data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "./cmd/atlas",
  ]
}

locals {
  env_content = file(".env")
}

locals {
  env_vars = {
    for line in split("\n", local.env_content) :
    split("=", line)[0] => split("=", line)[1]
    if length(split("=", line)) == 2 && line != "" && !startswith(line, "#")
  }
}

env "gorm" {
  url = "postgres://${local.env_vars.DB_USER}:${local.env_vars.DB_PASSWORD}@${local.env_vars.DB_HOST}:${local.env_vars.DB_PORT}/${local.env_vars.DB_NAME}?search_path=public&sslmode=disable"
  src = data.external_schema.gorm.url
  dev="docker://postgres/15/dev?search_path=public"
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
