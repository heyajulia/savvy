// FIXME: Calling this function 'std' causes a runtime error:
// "RUNTIME ERROR: can only index objects, strings, and arrays, got function."
local deno(package) =
  [package, 'https://deno.land/std@0.180.0/' + package];

local esm(package, version) =
  [package, 'https://esm.sh/v114/%s@%s' % [package, version]];

local dependencies = [
  deno('collections/'),
  deno('dotenv/'),
  esm('dedent', '0.7.0'),
  esm('luxon', '3.3.0'),
  esm('zod', '3.21.4'),
];

local permissions = '--allow-net=api.energyzero.nl,api.telegram.org --allow-read=.env,.env.defaults,.env.example --allow-env';

local commonArgs = permissions + ' --cached-only src/index.ts';

{
  // FIXME: I'd like something like JavaScript's Object.fromEntries() here.
  imports: {
    [dep[0]]: dep[1]
    for dep in dependencies
  },
  tasks: {
    compile: 'deno compile -o bot.exe ' + commonArgs,
    run: 'deno run ' + commonArgs,
    'run:prod': 'CHAT_ID=@energieprijzen ' + self.run,
  },
}
