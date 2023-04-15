local deno(name, file=name) =
  [name, 'https://deno.land/std@0.183.0/' + file];

local esm(package, version) =
  [package, 'https://esm.sh/v114/%s@%s' % [package, version]];

local dependencies = [
  deno('asserts', 'testing/asserts.ts'),
  deno('collections/'),
  deno('dotenv/'),
  esm('dedent', '0.7.0'),
  esm('luxon', '3.3.0'),
  esm('ramda', '0.29.0'),
  esm('zod', '3.21.4'),
];

local permissions = '--allow-net=api.energyzero.nl,api.telegram.org --allow-read=.env,.env.defaults,.env.example --allow-env';

local commonArgs = permissions + ' --cached-only src/index.ts';

{
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
