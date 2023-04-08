let std =
      λ(package : Text) →
        { mapKey = package
        , mapValue = "https://deno.land/std@0.180.0/${package}"
        }

let esm =
      λ(package : Text) →
      λ(version : Text) →
        { mapKey = package
        , mapValue = "https://esm.sh/v114/${package}@${version}"
        }

let permissions =
      "--allow-net=api.energyzero.nl,api.telegram.org --allow-read=.env,.env.defaults,.env.example --allow-env"

let commonArgs = "${permissions} --cached-only src/index.ts"

in  { imports =
      [ std "collections/"
      , std "dotenv/"
      , esm "dedent" "0.7.0"
      , esm "luxon" "3.3.0"
      , esm "zod" "3.21.4"
      ]
    , tasks =
      { compile = "deno compile -o bot.exe ${commonArgs}"
      , run = "deno run ${commonArgs}"
      }
    }
