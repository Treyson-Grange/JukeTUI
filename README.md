# IMPORTANT

This is legacy code before Spotify decided to deprecate the recommendation system. This branch is more out of hope than anything that Spotify brings the endpoint back.

See [Issue 16](https://github.com/Treyson-Grange/JukeTUI/issues/16) for more information

# JukeTUI

JukeTUI is a command-line interface (TUI) application that allows users to control Spotify, manage playback, and receive music recommendations through an intuitive jukebox system.

![JukeTUI App Preview](/screenshots/app-preview.png)

### Features:

- Jukebox: Get tailored song recommendations based on your currently playing track and easily queue songs for seamless listening.
- Library: Browse your Spotify music library, including albums and playlists, and play your favorite tracks directly from the app.
- Playback Bar: Effortlessly manage your music with controls to play, pause, skip tracks, and view what’s currently playing.

## Setup

JukeTUI requires the use of Spotify's API, and you will have to create your own "spotify app" on the Spotify dashboard for developers. This process isn't that bad. Details are discussed below

JukeTUI uses the Spotify Web API, which doesn't handle playback on its own. Therefore you will have to start your playback on a official spotify app, or a lightweight spotify app such as [spotifyd](https://github.com/Spotifyd/spotifyd). Once playback has started, you can change playback state, and change playlists/albums from JukeTUI.

### Connecting to Spotify's API

`JukeTUI` is specifically a spotify app. Therefore you will need to follow the steps to create your own spotify application. Don't worry, it's easy! As long as you are subscribed to spotify premium that is.

1. Go to the Spotify dashboard for developers
2. You will need to "Create app" and follow the instructions there.
3. In the settings of the new app, you will find a client ID and client secret
4. Copy `.env.example` into `.env` and paste your client ID and client secret into the corresponding variables.
5. You will then have to setup a Redirect URI. This is done in the app dashboard. click settings, Edit, and change the Redirect URIs and set it to `http://localhost:8080/callback`
6. On run, you will be asked to grant spotify permissions.
7. On return, you will be in the app, ready to go.

### Setup your environment

#### `.env`

- To set up your `.env`, copy `.env.example` into `.env`.

```
SPOTIFY_ID="{ From the developer dashboard }"
SPOTIFY_SECRET="{ From the developer dashboard }"
SPOTIFY_PREFERENCE="{ Either 'album' or 'playlist' }"
```

- Spotify ID and Cecret are for Spotify API auth
- Spotify Preference will alter what is displayed in the library. Your saved albums or your saved playlists.

## Use

To run JukeTUI, simply run `go run .`.

### Keybinds

General

- Quit: q

Library

- Navigate Library: Up/Down arrows
- Change Library page: Left/Right arrows
- Play selected library item: Enter

Playback

- Play/Pause: p
- Skip: n
- Toggle shuffle: s

Jukebox

- Get recommendation: r
- Add recommendation to queue: c

#### Custom Keybinds

Custom keybinds for most major functionality is available through changes in your environment file.

For available keybinds, see `.env.example`
