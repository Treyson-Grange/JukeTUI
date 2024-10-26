# JukeTUI

JukeTUI is a command-line interface (TUI) application that allows users to control Spotify, manage playback, and receive music recommendations through an intuitive jukebox system.

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

## Known Issues

- [ ] Emojis with 2 or 4 runes screw up our lipgloss formatting, lipgloss is aware of this issue, see [here](https://github.com/charmbracelet/lipgloss/issues/55)
- [ ] On window resize, UI breaks. (sometimes)
- [ ] Spotify Freemium does not functon with this app.

## IDEAS

If checked, it's gonna happen.

If not, it's a stretch goal.

- [x] Favorite system. You can favorite a playlist or album IN APP, and it will go to the TOP of your library list at all times, with a little star.

  - This would most likely be done by writing to a JSON file, and read on start.
  - Store: album/playlist name. album/playlist author. URI.
  - Needs a way to star albums, and a way to remove them (f for favorite? f again to toggle? or maybe we can do like a key combo. Who knows)

- [ ] FREE spotify system

  - Not everyone pays for spotify premium. This kills our API at points:
    - Pause, Skip, Start, Shuffle, Add to queue.
  - This kills a lot of the app.
  - We can check if someone has spotify premium, 403 PREMIUM_REQUIRED || /me endpoint, it returns "product" which if premium is "premium".
  - And according to forums, spotify free accounts can still create 'apps'
  - So, if we can set it up such that once we get our key, we can call a known "spotify premium exclusive" endpoint
  - Based on output, we can set some var, changing functionality in our application.
  - Functionality without spotify premium is SUPER bare bones. I'm pretty sure we won't be able to interact with the library.
    - We can GET the library, but we cannot PLAY the albums/playlists
    - We can GET a reccomendation, but we cannot add it to the queue.
    - We can LOOK at how much
  - Ideas to get around this (make it useful with possible endpoints):
    - Change library into a history section. /me/player/recently-played
    - Change Juke section into a queue section.
    - Disable Pause, skip, add to queue.

- [x] Audiobook / Podcast integration?

  - With minimal testing, I realized that playback state doesn't include podcast stuff.
  - Structures:
    - "shows"
      - These are podcast shows, I like to think of them as series.
      - They can be played with the /play endpoint
      - Starts from TOP of list unless offset
    - "episode"
      - These are episodes of the show.
      - Also playable with /play
    - "audiobook"
      - Interestingly, these are also shows, (spotify:show:0WAgAsP8MT7ae5p6gwL257 is an audio book)

- [ ] Liked songs?

  - Liked songs isn't a playlist, and won't show up on the library section. I think it would be NICE to have it available when playlists are being displayed.
  - `/me/tracks` is the endpoint that GETS them all, but there is no endpoint to PLAY them. It can get up to a total of 50.
  - First thoughts are to get total liked songs, randomize it ourselves. Online forums weren't super helful.

- [x] Scrollable Library.
  - Spotify API has a limit of 50 on the album/playlist endpoint.
  - But it also has a "offset". This offset can be used to create artificial pagination
    - Store current length, offset by that to go to next page.
    - Total pages = total playlists or albums / pageSize.
    - (change library to have header, say "\_\_\_'s Albums/Playlists | Page 2/6")
    - Store the next call, so there is no wait.
