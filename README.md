# JukeTUI

JukeTUI is a command-line interface (TUI) application that allows users to control Spotify, manage playback, and receive music recommendations through an intuitive jukebox system.

### Features:

- Jukebox: Get tailored song recommendations based on your currently playing track and easily queue songs for seamless listening.
- Library: Browse your Spotify music library, including albums and playlists, and play your favorite tracks directly from the app.
- Playback Bar: Effortlessly manage your music with controls to play, pause, skip tracks, and view what’s currently playing.

## Setup

JukeTUI requires the use of Spotify's API, and you will have to create your own "spotify app" on the Spotify dashboard for developers. This process isn't that bad. Details are discussed below

### Connecting to Spotify's API

`JukeTUI` is specifically a spotify app. Therefore you will need to follow the steps to create your own spotify application. Don't worry, it's easy! As long as you are subscribed to spotify premium that is.

1. Go to the Spotify dashboard for developers
2. You will need to "Create app" and follow the instructions there.
3. In the settings of the new app, you will find a TUIent ID and TUIent secret
4. Copy `.env.example` into `.env` and paste your TUIent ID and TUIent secret into the corresponding variables.
5. You will then have to setup a Redirect URI. This is done in the app dashboard. TUIck settings, Edit, and change the Redirect URIs and set it to `http://localhost:8080/callback`
6. On run, you will be asked to grant spotify permissions.
7. On return, you will be in the app, ready to go.

### Setup your environment

#### `.env`

- To set up your `.env`, copy `.env.example` into `.env`.

```
SPOTIFY_ID={{ From the developer dashboard }}
SPOTIFY_SECRET={{ From the developer dashboard }}
SPOTIFY_PREFERENCE={{ Either 'album' or 'playlist' }}
```

- Spotify Preference will alter what is displayed in the cursor section. Your saved albums or your saved playlists.
  - (This feature might turn into 'tabs' eventually, so we can have both)

## TODO

- [x] General UI
- [x] Better and more generic function calls
- [x] Reccomendations system (get/add to queue)
- [ ] Jukebox UI
  - [ ] Figure out general way to approach this. As window resizes, a jukebox ascii art would break.
  - [ ] So maybe we just have like a layout that LOOKS like a jukebox where it like has a disc and changes it
  - [ ] Or maybe we replace it with something else when window size gets too small.
  - [ ] Maybe a little animation that plays when you ask for a recc.
- [x] Fix errors when currently no playback, shouldn't exit
- [x] Playback BAR
  - [ ] Pause/Play that isn't ugly.
  - [ ] Progress in Seconds
  - [ ] Image maybe? I feel like ive seen something like this, image in terminal out of unicode chars.
  - [ ] Steal inspo from spotify, but simpler, as keybinds are used to do most things.
    - [ ] Shuffle Display
    - [ ] Song title, artist,
    - [ ] Track progress 0:54 / 1:52
- [x] Playlists BOX
  - [x] List all albums/playlists (see below)
  - [x] Cursor to roll over them.
  - [x] Button to play them.
- [x] I am a album guy, some people are playlist people. Therefore, we'll need a toggle to either display Users saved albums or users playlists.

  - [x] Decide how we want to toggle this. In a config? in env? with a arg? probably env. Lets do env
  - [x] Add new env variable to both env and example
  - [x] This will run on start, so we can have a homescreen
  - [ ] I want a tab secion, so I can swithc tabs and see my playlists or my albums.
  - [ ] I think lipgloss has tabs, experiment with them

- [ ] API key lasts for 1 hour. We need a system to get new ones.
  - [x] Base functionality
  - [ ] Testing: If we lock our computer but come back before it ends, will it still correctly refresh at the right time?
  - [ ] Testing: What if we don't come back?

## Known Issues

- [ ] Emojis with 2 or 4 runes screw up our lipgloss formatting, lipgloss is aware of this issue, see [here](https://github.com/charmbracelet/lipgloss/issues/55)
- [ ] On window resize, UI breaks.
- [ ]

## IDEAS :)

- Favorite system. You can favorite a playlist or album IN APP, and it will go to the TOP of your library list at all times, with a little star.
  - This would most likely be done by writing to a JSON file, and read on start.
  - Store: album/playlist name. album/playlist author. URI.
  - Needs a way to star albums, and a way to remove them (f for favorite? f again to toggle? or maybe we can do like a key combo. Who knows)
