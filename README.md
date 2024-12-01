## What is this?

These are [kindrobot](https://kindrobot.ca/)'s dot files with a focus on
idempotent units. This is a work in progress. To use, clone this repo, and run
any or all of units, e.g. `./units/install-apt-cli-pkgs`.

## Philosophy

All units are idempotent. This means you should be able to run units as many
times as you fancy. If a unit updates (e.g. you git pull with changes to a unit
you've already run), you can run the unit again, and it should idempotently
apply updates/new changes.

## License

[MDGPL](./LICENSE)

