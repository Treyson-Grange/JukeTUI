# JukeCLI

## Idea

-   Jukebox on your command line.
-   Spotify integration
-   playback controls.
-   Jukebox section that gives reccs or random songs?
-   Idea is loose for now, as I think about what is actually useful.
-   Slick UI

## State of the project

-   This is so far just a proof of concept. I have done enough work that I can login and access user data

## Connecting to Spotify's API

`JukeCLI` is specifically a spotify command line interface. Therefore you need to follow these steps for JukeCLI to run.

1. Go to the Spotify dashboard for developers
2. You will need to "Create app" and follow the instructions there.
3. In the settings of the new app, you will find a Client ID and Client secret
4. Copy `.env.example` into `.env` and paste your Client ID and Client secret into the corresponding variables.
5. You will then have to setup a Redirect URI. This is done in the app dashboard. Click settings, Edit, and change the Redirect URIs and set it to `http://localhost:8080/callback`
6. On run, you will be asked to grant spotify permissions.
7. On return, you will be in the app, ready to go.
