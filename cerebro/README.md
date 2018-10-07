# 🧠 Cerebro

Cerebro is the application that finds characters, character sources, and character issues from external sources and imports those resources as local resources into the appearances database so that all a characters.

## CLI Commands

- `cerebro import [resource]`: Imports external resources as local resources. Available resources: `characters`, `charactersources`, `characterissues`
- `cerebro start characterissues`: Starts a long-running process that consumes messages from a queue and imports character issues. Essentially, this command consumes from a queue + imports characterissues. Send a `SIGINT`/`ctrl` + `c` to cleanly quit the process.

## What counts as an appearance

`characterissue.go` contains the logic for aggregating a character's issues and counting it as an appearance and persisting it. 

There are two categories for the type of appearances: **main** and **alternate** appearances.

Main appearances refer to the original character (for DC the rules are a little different). Alternate appearances refer to a character's alternate reality appearances, such as the Marvel Ultimate line.

Here are the rules for what counts as an appearance.

What does NOT count:
- Variants
- Reprints, such as 2nd printings or issues reprinted in another language.
- Flip-books, ashcans, magazines, trade paperbacks, etc.

**Marvel**:

Main
- The character's 616 counterpart makes an appearance.
- The character's 616 counterpart was transported into another reality, such as House of M or Age of X.

Alternate
- The character appears, such as Spider-Man 2099, but in a difference universe, such as the 2099 or Ultimate line.

**DC**:

Main
- Due to DC's nature, almost all appearance are main appearances, except comics about other mediums, such as TV shows, video games, movies, etc. 

Alternate
- Comics involving tv shows, video games, and movies are alternate appearances.
