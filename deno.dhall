let std = λ(package : Text) → "https://deno.land/std@0.180.0/${package}"

let esm =
      λ(package : Text) →
      λ(version : Text) →
        "https://esm.sh/v114/${package}@${version}"

let permissions =
      "--allow-net=api.energyzero.nl,api.telegram.org --allow-read=.env,.env.defaults,.env.example --allow-env"

let indexFile = "src/index.ts"

in  { imports =
      { collections/ = std "collections/"
      , dotenv/ = std "dotenv/"
      , dedent = esm "dedent" "0.7.0"
      , luxon = esm "luxon" "3.3.0"
      , zod = esm "zod" "3.21.4"
      }
    , tasks =
      { compile = "deno compile -o bot.exe ${permissions} ${indexFile}"
      , run = "deno run ${permissions} ${indexFile}"
      }
    }
