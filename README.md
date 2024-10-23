# JukeCLI

## Idea

- Jukebox on your command line.
- Spotify integration
- playback controls.
- Jukebox section that gives reccs or random songs?
- Idea is loose for now, as I think about what is actually useful.
- Slick UI

## State of the project

- This is so far just a proof of concept. I have done enough work that I can login and access user data
- Little update, its set up with bubble tea now, so UI and general flow of data and IO should be easier
- There is now functionality to get playback state, display it, pause/play playback, and skip tracks.
- Currently we are doing the same http request code, so TODO: fix that.

## Connecting to Spotify's API

`JukeCLI` is specifically a spotify command line interface `. Therefore you need to follow these steps for JukeCLI to run.

1. Go to the Spotify dashboard for developers
2. You will need to "Create app" and follow the instructions there.
3. In the settings of the new app, you will find a Client ID and Client secret
4. Copy `.env.example` into `.env` and paste your Client ID and Client secret into the corresponding variables.
5. You will then have to setup a Redirect URI. This is done in the app dashboard. Click settings, Edit, and change the Redirect URIs and set it to `http://localhost:8080/callback`
6. On run, you will be asked to grant spotify permissions.
7. On return, you will be in the app, ready to go.

## KEEP THESE IN MIND

- API key lasts for 1 hour. We need a system to get new ones.

## TODO

- [ ] General UI
- [ ] Better and more generic function calls
- [x] Reccomendations system (get/add to queue)
- [ ] Jukebox UI (prettyyyyy)
- [x] Fix errors when currently no playback, shouldn't exit
- [ ] Playback BAR
  - [ ] Pause/Play.
  - [ ] Progress in Seconds
  - [ ] Image maybe? I feel like ive seen something like this, image in terminal out of unicode chars.
- [ ] Playlists BOX
  - [ ] List all albums/playlists (see below)
  - [ ] Cursor to roll over them.
  - [ ] Button to play them.
- [ ] I am a album guy, some people are playlist people. Therefore, we'll need a toggle to either display Users saved albums or users playlists.
  - [x] Decide how we want to toggle this. In a config? in env? with a arg? probably env. Lets do env
  - [x] Add new env variable to both env and example
  - [ ] Read in env, and in the T case (for testing for now) we will do different things based on it.
  - [ ] This will run on start, so we can have a homescreen
