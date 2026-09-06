# JukeTUI Interview Preparation Guide

## 1. Project Overview

### What JukeTUI Does

JukeTUI is a **terminal user interface (TUI) for controlling Spotify**. It's a standalone application you run in your terminal that lets you browse your Spotify library (albums or playlists), play/pause music, skip tracks, shuffle, and view the queue—all without opening the Spotify app. It displays your album art as pixelated, colored blocks in the terminal.

### The Problem It Solves

The Spotify Web API doesn't have built-in playback capabilities (it can only control already-playing music). For many developers, opening the full Spotify application is overkill just to browse and control music. JukeTUI provides a lightweight, terminal-native interface for music control, appealing to developers and power users who live in the terminal.

### Main User Flow

1. **Login**: User runs the app → browser opens → user authenticates with Spotify → authorization code is returned to the app
2. **Browse**: User sees their albums or playlists in a paginated list in the terminal
3. **Play**: User selects an album/playlist with Enter → music starts playing
4. **Control**: User can pause/play (p), skip (n), shuffle (s), mark favorites (f), navigate with arrow keys
5. **Discover**: Album art displays as colored characters; queue shows next 5 tracks

### Major Technologies & APIs

- **Go 1.23.2**: Programming language
- **Bubbletea** (`github.com/charmbracelet/bubbletea`): TUI framework providing event loop and component model
- **Lipgloss** (`github.com/charmbracelet/lipgloss`): Styling library for terminal layouts
- **Spotify Web API**: RESTful API for accessing user library and controlling playback
- **OAuth 2.0**: Authorization flow via `https://accounts.spotify.com/authorize` and token endpoint
- **Image Processing**: `golang.org/x/image` for resizing album art to fit terminal

### High-Level Architecture

```
User Input (Keyboard)
        ↓
    [Bubbletea Event Loop] ← Models (tea.Msg)
        ↓
  [Update Handler]
   ├─ Routes input to handlers
   ├─ Updates Model state
   └─ Returns Commands (async operations)
        ↓
  [Commands Execute Asynchronously]
   ├─ Spotify API requests (HTTP)
   ├─ Image fetching & processing
   └─ Timer/scheduling commands
        ↓
  [Messages Sent Back to Event Loop]
        ↓
  [View Renders Model to Terminal]
```

### What Makes It Technically Interesting

1. **Spotify OAuth Integration**: Implements proper OAuth2 flow with local HTTP callback server
2. **Async TUI Architecture**: Separates I/O from rendering using Bubbletea's command/message pattern
3. **Image to Terminal Display**: Converts album images to ANSI colored blocks by sampling pixels
4. **Pagination & State Management**: Handles pagination of large Spotify libraries with favorites system
5. **Token Lifecycle Management**: Attempts to refresh OAuth tokens before expiration
6. **Real-time Updates**: Polls Spotify playback state at 2-second intervals

---

## 2. Architecture Walkthrough

### Entry Points

- **[main.go](src/main.go) - `main()` function (line 307-335)**
  - Loads `.env` file for configuration
  - Initiates OAuth login flow
  - Loads favorites from JSON files
  - Creates initial Model
  - Launches Bubbletea program

### Major Modules & Responsibilities

#### Authentication ([spotifyAuth.go](src/spotifyAuth.go))

- `OpenLoginPage()`: Opens Spotify auth URL in default browser
- `GetCodeFromCallback()`: Runs local HTTP server on `:8080` to capture authorization code
- `GetSpotifyToken()`: Exchanges code for access token via Spotify token endpoint
- `RefreshSpotifyToken()`: Uses refresh token to get new access token

**Key insight**: The HTTP server started in `GetCodeFromCallback()` has no cleanup mechanism—it keeps listening forever.

#### API Communication Layer

**[spotifyRequests.go](src/spotifyRequests.go)** - Generic HTTP wrapper

- `genericRequest[T]()`: Generic HTTP request builder (GET/POST/PUT)
- `genericFetch[T]()`: GET wrapper
- `genericPut()`: PUT wrapper
- `genericPost()`: POST wrapper
- Uses Go generics for type-safe response handling

**[spotifyHandlers.go](src/spotifyHandlers.go)** - High-level API operations

- `handleGenericFetch/Put/Post()`: Error-checking wrappers
- `handleFetchPlayback()`: Gets current playback state
- `handleFetchLibrary()`: Fetches albums/playlists with favorites filtering
- `handleGetQueue()`: Fetches next 5 queued tracks
- Each returns `tea.Cmd` for async execution

**[spotifyUtil.go](src/spotifyUtil.go)** - Utilities

- `createEndpoint()`: Builds Spotify API URLs with query parameters
- `CheckTokenExpiryCmd()`: Checks if token expired, triggers refresh if needed

#### UI Rendering ([uiUtil.go](src/uiUtil.go))

- `getUiElements()`: Orchestrates all UI components
- `getLibText()`: Renders album/playlist list with cursor, tabs, pagination
- `getPlayBack()`: Renders playback bar (track info, duration, shuffle state)
- `getVisualQueue()`: Renders next 5 tracks in queue
- `getTabs()`: Renders active/inactive album vs playlist tabs
- Image handling:
  - `fetchImage()`: HTTP GET album art URL
  - `resizeImage()`: Resizes image to terminal dimensions
  - `makeNewImage()`: Orchestrates fetching + resizing + rendering
  - `printImage()`: Converts pixels to ANSI colored text

#### State Management ([models.go](src/models.go))

- **Model struct**: Single source of truth
  - Playback state (current track, device ID, shuffle, progress)
  - OAuth tokens and expiry time
  - Library list (albums/playlists), cursor, pagination offset
  - Favorites (separate for albums and playlists)
  - Queue, error message, loading flag
- **Message Types**: `PlaybackState`, `SpotifyAlbum`, `SpotifyPlaylist`, `Queue`, custom `playbackMsg`, `progressMsg`

#### Data Persistence ([json.go](src/json.go))

- `readJSON()`: Reads favorites from `data/favorites/albums.json` or `data/favorites/playlists.json`
- `writeJSONFile()`: Adds a new favorite
- `removeFromJSON()`: Removes a favorite by matching title/author/URI

#### Utilities ([util.go](src/util.go), [log.go](src/log.go))

- `setKeybinds()`: Reads keybinds from `.env`
- `scheduleNextFetch()`: Returns a command that sleeps and sends `playbackMsg`
- `scheduleProgressInc()`: Returns a command that sleeps and sends `progressMsg`
- Logging to `data/logs/errors.log` and `data/logs/info.log` (only in DEVELOPMENT mode)

### How State Flows

```
[OAuth Login]
      ↓
[Get Access Token + Refresh Token]
      ↓
[Initialize Model with token + load favorites]
      ↓
[Init() commands launch]:
   - Fetch playback state
   - Fetch library total count
   - Schedule progress increment
   - Fetch library with pagination
   - Fetch queue
      ↓
[Event Loop Processes Messages]
   - User keyboard input → calls handlers → returns commands
   - Playback state update → updates UI
   - Library/queue update → updates display
      ↓
[View() re-renders terminal]
```

### How Spotify API is Accessed

1. **Authentication**: OAuth2 code → token via `https://accounts.spotify.com/api/token`
2. **Requests**: All subsequent requests use `Authorization: Bearer {token}` header
3. **Endpoints used**:
   - `/me/albums` → Fetch user's saved albums
   - `/me/playlists` → Fetch user's playlists
   - `/me/player` → Get playback state
   - `/me/player/play`, `/me/player/pause`, `/me/player/next` → Control playback
   - `/me/player/shuffle` → Toggle shuffle
   - `/me/player/queue` → Get queue
4. **Error handling**: Minimal—logs errors, returns empty structs

### Authentication Workflow Detailed

```
1. main() calls OpenLoginPage(clientID)
   ↓
2. Browser opens:
   https://accounts.spotify.com/authorize?client_id=...&scope=user-read-private+user-modify-playback-state+...
   ↓
3. User logs in and grants permissions
   ↓
4. Spotify redirects to http://127.0.0.1:8080/callback?code=AUTH_CODE
   ↓
5. GetCodeFromCallback() HTTP handler captures code
   ↓
6. Browser closes (via JavaScript)
   ↓
7. GetSpotifyToken() POSTs to Spotify token endpoint with:
   - grant_type=authorization_code
   - code={AUTH_CODE}
   - client_id & client_secret (Basic auth)
   ↓
8. Returns: access_token (1 hour), refresh_token (infinite), expires_in (3600)
```

### UI/Input Handling

- Bubbletea captures keyboard events in `Update(msg tea.Msg)`
- Keybinds are configurable via `.env` (defaults: q=quit, p=play/pause, n=skip, s=shuffle, f=favorites)
- Arrow keys are hardcoded for list navigation
- `View()` is called after every message, re-renders entire terminal

### Async/Concurrency Model

- Bubbletea runs a single-threaded event loop
- Commands return `tea.Cmd` which are non-blocking goroutines
- All Spotify API calls run in goroutines spawned by commands
- No explicit synchronization (mutexes) needed because Model updates only happen in Update() (single-threaded)
- Playback state fetches every 2 seconds via `scheduleNextFetch()`
- Progress bar increments every 1 second via `scheduleProgressInc()`

### Configuration/Environment Handling

- `.env` file contains:
  - `SPOTIFY_ID`, `SPOTIFY_SECRET`: OAuth credentials
  - `SPOTIFY_PREFERENCE`: "album" or "playlist" (default display)
  - Custom keybinds: `QUIT`, `PLAYPAUSE`, `SKIP`, `SHUFFLE`, `FAVORITES`
  - `DEVELOPMENT`: Enable/disable logging to files
- Loaded via `godotenv` package

### Error Handling

- **HTTP errors**: `handleGenericFetch()` logs to `errorLogger`, returns empty struct
- **JSON parsing errors**: Logged, function returns false
- **Token expiry**: `CheckTokenExpiryCmd()` checks before each update
- **Offline**: If Spotify API is unreachable, UI shows last known state
- **No error modal**: Users see stale data or error messages in `Model.errMsg`

### Simple ASCII Architecture Diagram

```
                    ┌─────────────────────────────────┐
                    │    Spotify Web API              │
                    │  (OAuth, Albums, Playlists,     │
                    │   Playback, Queue)              │
                    └────────────┬────────────────────┘
                                 │
                ┌────────────────┴───────────────────┐
                │                                    │
           [spotifyAuth.go]              [spotifyRequests.go]
         (OAuth flow, tokens)            (HTTP GET/PUT/POST)
                │                                    │
                └────────────┬───────────────────────┘
                             │
                    [spotifyHandlers.go]
                 (High-level API operations)
                        ├─ fetch playback
                        ├─ fetch library
                        ├─ fetch queue
                        └─ control playback
                             │
                    ┌────────┴─────────┐
                    │                  │
          [Bubbletea Event Loop]     [json.go]
          (Model.Update/View)    (Favorites persistence)
                    │                  │
                ┌───┴──────────────────┴─────┐
                │                            │
            [uiUtil.go]            [util.go, log.go]
          (Render terminal)       (Keybinds, logging,
           ├─ Library list         scheduling)
           ├─ Playback bar
           ├─ Album art
           └─ Queue
                │
            [Terminal Output]
```

---

## 3. Important Execution Flows

### Application Startup

**File**: [main.go](src/main.go#L307-L335)

```
1. godotenv.Load(".env") → Load environment variables

2. OpenLoginPage(clientID) →
   - Build Spotify auth URL with scopes
   - xdg-open (Linux) / open (macOS) / rundll32 (Windows) to open browser

3. GetCodeFromCallback() →
   - Start HTTP server on :8080
   - Block until authorization code received via callback
   - Parse code from URL query parameter
   - Return to browser with success message

4. GetSpotifyToken(clientID, clientSecret, code) →
   - POST to https://accounts.spotify.com/api/token
   - Headers: Basic auth with client_id:client_secret
   - Body: grant_type=authorization_code, code=..., redirect_uri=...
   - Parse response: get access_token, refresh_token, expires_in

5. readJSON("data/favorites/albums.json") + readJSON("data/favorites/playlists.json") →
   - Load favorite albums/playlists from local JSON files

6. initialModel(token, "album", favoriteAlbums, favoritePlaylists) →
   - Create Model struct with all state initialized
   - Set tokenExpiresAt = now + 1 hour

7. tea.NewProgram(model).Run() →
   - Enter Bubbletea event loop
   - Call model.Init() to get initial commands
```

### Init() - Initial Commands

**File**: [main.go](src/main.go#L36-L45)

Returns batch of concurrent commands:

1. `handleFetchPlayback()` → Fetch current playback state
2. `handleGetLibraryTotal()` → Fetch total album/playlist count
3. `scheduleProgressInc()` → Start progress bar timer
4. `handleFetchLibrary()` → Fetch first page of library
5. `handleGetQueue()` → Fetch next 5 tracks

All run asynchronously, results come back as messages to `Update()`.

### User Presses 'p' (Play/Pause)

**File**: [main.go](src/main.go#L50-L61)

```
1. Update(tea.KeyMsg) detects key "p"

2. If m.state.IsPlaying:
   - Call handleGenericPut("/me/player/pause", token, nil, nil)
   Else:
   - Call handleGenericPut("/me/player/play", token,
       nil, {"device_id": m.state.Device.ID})

3. handleGenericPut() (in spotifyHandlers.go)
   - Wraps genericPut() with error logging

4. genericPut() (in spotifyRequests.go)
   - Calls genericRequest[struct{}]() with PUT method
   - Constructs HTTP request with Authorization header
   - POSTs to Spotify endpoint
   - Returns status code and error

5. Command completes
   - No message returned (playback state will be fetched in next 2-second cycle)
   - Return m, nil (no state change)
```

### User Presses 'Enter' (Select Album/Playlist)

**File**: [main.go](src/main.go#L174-L187)

```
1. Update(tea.KeyMsg) detects "enter"

2. If not currently playing:
   - Return m, nil (do nothing)

3. If listDetail == "album":
   - Call handleGenericPut("/me/player/shuffle", token,
       {"state": "false"}, nil)
   Else (playlist):
   - Call handleGenericPut("/me/player/shuffle", token,
       {"state": "true"}, nil)

4. Call handleGenericPut("/me/player/play", token,
     {"device_id": m.state.Device.ID},
     {"context_uri": m.libraryList[m.cursor].uri})
   - This tells Spotify to play the selected album/playlist

5. Spotify starts playing from that context
   - Playback state fetches every 2 seconds will pick up the new song

6. View() re-renders with new track info
```

### User Presses 'Right Arrow' (Next Page)

**File**: [main.go](src/main.go#L138-L154)

```
1. Update(tea.KeyMsg) detects "right"

2. m.loading = true (show loading indicator)

3. Calculate next offset:
   - If not at end: offset += (height - UI_LIBRARY_SPACE - favorites_count)
   - If at end: offset = 0 (wrap around)

4. Call handleFetchLibrary(favorites, token, listDetail, height, offset)
   - This command fetches next page from Spotify API

5. In handleFetchLibrary():
   - Fetch limit = min(height, 50)
   - If album: GET /me/albums?limit=50&offset=...
   - If playlist: GET /me/playlists?limit=50&offset=...
   - Filter out favorites from results
   - Return SpotifyAlbum or SpotifyPlaylist message

6. Update() receives SpotifyAlbum/SpotifyPlaylist message:
   - Clear m.libraryList
   - Add favorites to front of list
   - Add fetched items to list
   - Set m.loading = false
   - Return m, nil

7. View() renders new page
```

### User Presses 'f' (Toggle Favorite)

**File**: [main.go](src/main.go#L86-L120)

```
1. Update(tea.KeyMsg) detects "f"

2. file = "data/favorites/albums.json" (or playlists.json based on m.listDetail)

3. Check if current item is already in favorites:
   - If found:
     - removeFromJSON(file, oldFavorite)
     - m.favoriteAlbums/favoritePlaylists = readJSON(file)
   - If not found:
     - writeJSONFile(file, LibraryFavorite{name, artist, uri})
     - m.favoriteAlbums/favoritePlaylists = readJSON(file)

4. Call handleFetchLibrary() with updated favorites
   - This re-fetches the page so favorites move to top

5. View() shows updated list with heart icon next to favorites
```

### Playback State Update (Every 2 Seconds)

**File**: [main.go](src/main.go#L200-L220), [util.go](src/util.go#L11-L19)

```
1. scheduleNextFetch(2*time.Second) command executes:
   - time.Sleep(2 seconds)
   - Return playbackMsg{}

2. Update(playbackMsg) detects the message:
   - Call handleFetchPlayback(m.token)
   - This command fetches current playback state

3. handleFetchPlayback returns tea.Cmd:
   - Calls handleGenericFetch[PlaybackState]("/me/player", token, nil, nil)
   - Returns PlaybackState message

4. Update(PlaybackState) receives state:
   - If track changed (m.state.Item.Name != msg.Item.Name):
     - Call makeNewImage(imageURL) to fetch and process album art
     - Set m.image = new image
     - Update m.state = msg
     - Call CheckTokenExpiryCmd() to refresh token if needed
     - Return handleGetQueue() to update queue display
   - Else:
     - Update progress if > 1 second different
     - Update m.state = msg

5. View() re-renders with updated track info and image
```

### Token Expiry Refresh

**File**: [spotifyUtil.go](src/spotifyUtil.go#L27-L33)

```
1. Every Update() cycle calls CheckTokenExpiryCmd(m)

2. CheckTokenExpiryCmd() checks:
   - if time.Now().After(m.tokenExpiresAt):
     - Return refreshSpotifyTokenCmd()
   - Else:
     - Return nil

3. If token expired, refreshSpotifyTokenCmd() executes:
   - Call RefreshSpotifyToken(refreshToken, clientID, clientSecret)
   - POST to https://accounts.spotify.com/api/token with:
     grant_type=refresh_token
     refresh_token={old_refresh_token}
     (with Basic auth of client_id:secret)

4. Spotify returns new access_token and expires_in

5. Update(SpotifyTokenResponse) receives response:
   - m.token = new access token
   - m.tokenExpiresAt = now + expires_in
   - Return m, nil

6. Subsequent Spotify API calls use new token
```

### Image Fetching and Display

**File**: [uiUtil.go](src/uiUtil.go#L100-L165)

```
1. When track changes, makeNewImage(imageURL, width, height) called

2. fetchImage(url) executes:
   - HTTP GET the album art URL
   - image.Decode() auto-detects format (PNG/JPG/GIF)
   - Return decoded image.Image

3. resizeImage(img, width, height):
   - Create new RGBA image at target size
   - Use Catmull-Rom interpolation (high quality)
   - Scale original image into new buffer

4. printImage(img) converts to terminal:
   - For each pixel in resized image:
     - Get RGBA color
     - Convert to ANSI 256-color code: \x1b[48;2;R;G;Bm
     - Write "  " (two spaces) with that background color
   - Result: ~16-32 colored blocks per image

5. Return string of ANSI codes

6. View() renders this string in the jukebox panel
```

---

## 4. Best Code to Show in Interview

### 1. OAuth 2.0 Implementation ([spotifyAuth.go](src/spotifyAuth.go#L46-L73))

**Functions**: `GetSpotifyToken()`, `RefreshSpotifyToken()`

**What it does**:

- Implements proper OAuth2 authorization code flow with Spotify
- Exchanges authorization code for access token
- Handles token refresh for automatic re-authentication

**Why it's interesting**:

- Shows understanding of OAuth2 security (client ID/secret in Basic auth header)
- Proper use of `http.NewRequest()` and `net/url` for encoding
- Uses Basic Authentication correctly: `req.SetBasicAuth()`
- Token refresh logic prevents user re-login during long sessions
- Error handling with context (though minimal)

**What to say**:
"Here's how we handle Spotify authentication. We use OAuth2 with the authorization code grant type. First, we build an auth URL with scopes we need—in this case, read playback state and modify playback state. We open it in the browser, and the user logs in. Spotify redirects to our local callback server on port 8080 with an authorization code. We then exchange that code for an access token by POSTing to Spotify's token endpoint with our client credentials in the Authorization header using Basic auth. The token lasts 1 hour, so we also save the refresh token so the app can automatically get a new token without asking the user to log in again."

**Likely questions**:

- Q: "Why do you pass client credentials in Basic auth instead of the request body?"
  A: "The Spotify API spec requires it for this flow. Using Basic auth keeps the credentials separate from the request body and is more secure since Basic auth often gets stripped from logs."

- Q: "What if token refresh fails?"
  A: "Currently we return an error and log it, but we don't have fallback logic. In production, we should probably show the user a message or start the login flow again."

- Q: "Why do you save both access_token and refresh_token?"
  A: "The access token is what we use for every API request, but it only lasts 1 hour. The refresh token is long-lived and lets us get a new access token without bugging the user to log in again."

---

### 2. Generic HTTP Request Handler with Generics ([spotifyRequests.go](src/spotifyRequests.go#L12-L58))

**Function**: `genericRequest[T any]()`

**What it does**:

- Single generic function handles GET/PUT/POST to Spotify API
- Constructs requests with proper headers and authentication
- Parses JSON responses into arbitrary type T
- Returns status codes for debugging

**Why it's interesting**:

- Demonstrates Go 1.18 generics (parametric polymorphism)
- Good separation of concerns: HTTP handling separate from business logic
- Type-safe: compiler ensures response is parsed into correct struct
- Reduces code duplication (would otherwise need 3 versions for GET/PUT/POST)
- Shows error handling with multiple failure points

**What to say**:
"This is the core of our API communication layer. Instead of writing separate functions for each endpoint, we use Go generics to write one generic `genericRequest` function that can handle any response type. The function takes a method (GET/PUT/POST), the endpoint path, and the access token. It builds the full Spotify API URL, adds the Bearer token in the Authorization header, and makes the HTTP request. For GET requests, we parse the JSON response into whatever type the caller specifies—T can be PlaybackState, SpotifyAlbum, anything. For PUT/POST, we just return the status code. This keeps all the HTTP boilerplate in one place and makes adding new API endpoints really easy."

**Likely questions**:

- Q: "Why did you choose to use generics here instead of just using json.Unmarshal with interface{}?"
  A: "Generics give us type safety. With interface{}, we'd lose the type at compile time and could end up putting the wrong type in a map. With generics, the compiler verifies that the caller uses the right type."

- Q: "How do you handle errors from the Spotify API?"
  A: "We check the HTTP status code. If we get 403 (Forbidden), that's a permissions issue. Otherwise we try to parse the response. Errors get returned to the caller. The handler functions log them."

- Q: "What's the timeout for your HTTP client?"
  A: "20 seconds. That's arbitrary but prevents the app from hanging forever if the network is slow. For a TUI app, we might want to make it configurable so users on slow networks don't timeout."

---

### 3. Pagination and Favorites Filtering ([spotifyHandlers.go](src/spotifyHandlers.go#L59-L106))

**Function**: `handleFetchLibrary()`

**What it does**:

- Fetches user's albums or playlists from Spotify API with pagination
- Filters out items that are already in user's favorites
- Ensures favorites appear at the top of the list
- Handles the limit correctly when favorites have already been fetched

**Why it's interesting**:

- Shows state management complexity: must coordinate between API responses and local state
- Demonstrates pagination logic (offset, limit, handling wraparound)
- Shows why favorites need special handling (user-created list, not from Spotify)
- The filtering logic is a bit complex/fragile—good talking point for improvements

**What to say**:
"This function does the heavy lifting for browsing your library. We need to show favorites at the top but still respect pagination. The trick is that when we fetch from Spotify, we don't know which items are favorites until we parse the response. So we first fetch a page, then filter out any items that match a favorite, then re-fetch with a larger offset to get the same number of non-favorite items. We put favorites at the front of the list. The issue here is it's inefficient—we're potentially making the same API call twice if there are favorites on the first page."

**Likely questions**:

- Q: "Why do you re-fetch the library when there are favorites?"
  A: "To ensure we show `height` items total. If we fetched 50 albums but 10 of them are already in favorites, we'd filter them out and show only 40 items. By re-fetching with a higher offset, we get the 10 we need."

- Q: "This seems inefficient. How would you improve it?"
  A: "We could cache which albums are favorites in a set so lookups are O(1) instead of O(n). Or we could fetch a larger initial batch to account for favorites. Better yet, we could modify the API calls to not include favorites in the response if the API supported it."

- Q: "What if favorites are on the second page?"
  A: "That's a problem we don't handle well right now. If you favorite an album on page 5, it should appear at the top. But we'd have to re-fetch all pages to find it, which is expensive. This is a real UX issue."

---

### 4. Bubbletea Event Loop and State Management ([main.go](src/main.go#L48-L287))

**Functions**: `Update()` method, `View()` method, `Model` struct in [models.go](src/models.go)

**What it does**:

- Implements the Bubbletea TUI framework's Update/View pattern
- Routes all user input and async messages through a single state machine
- Maintains all application state in a single Model struct
- Renders the entire UI based on Model state

**Why it's interesting**:

- Demonstrates functional reactive programming pattern (Elm architecture)
- Shows how async operations return messages that feed back into the event loop
- Type-safe message handling with Go's type switch
- Every interaction goes through Update() which makes behavior predictable
- Single source of truth (Model) simplifies debugging

**What to say**:
"The core of JukeTUI is built on Bubbletea, which uses an Elm-inspired architecture. Every interaction—keyboard input, API response, timer tick—becomes a message. The Update() function is the state machine that handles all messages. When a user presses a key, we switch on it and decide what to do. If we need to fetch data, we return a Command—an async operation. When that operation completes, it sends a message back, and Update() runs again. The View() function just renders the current Model to the terminal. This architecture is really clean because there's a single path for state changes, so you can reason about what happens when. Everything goes through Update()."

**Likely questions**:

- Q: "What's the difference between a Command and a Message in Bubbletea?"
  A: "A Message is data that comes into the system (user input, API response, timer). A Command is a function that performs work and returns Messages. Commands run asynchronously, so we can do I/O without blocking the UI."

- Q: "How do you handle errors from API calls?"
  A: "They're returned as error messages from commands. The Update() function has a case for `error` that sets m.errMsg and schedules the next fetch. It's not ideal—errors just get logged and the UI might show stale data."

- Q: "Why do you need to return a new Model from Update()?"
  A: "Bubbletea expects Update() to be a pure function that takes Model and Msg and returns a new Model. This immutability makes it easier to reason about state transitions."

---

### 5. Album Art Image Processing ([uiUtil.go](src/uiUtil.go#L95-L165))

**Functions**: `makeNewImage()`, `resizeImage()`, `printImage()`

**What it does**:

- Fetches album art image from URL
- Resizes to fit terminal dimensions
- Converts image pixels to ANSI colored characters
- Creates a pixelated album art display in the terminal

**Why it's interesting**:

- Shows low-level image manipulation (pixel sampling, color quantization)
- Demonstrates HTTP client usage for binary data
- Converts graphical data to text output (creative problem-solving)
- Uses `golang.org/x/image` for high-quality image scaling
- Understanding of ANSI escape codes for 256-color terminal output

**What to say**:
"Getting album art to display in a terminal is tricky. We start by fetching the image from Spotify's URL. Then we resize it using Catmull-Rom interpolation so it fits our terminal window. Here's the creative part: we can't draw pixels in a terminal, so we convert each pixel to an ANSI escape code that sets the background color of that cell. Each pixel becomes a two-character block with that color. The result is a pixelated but recognizable album cover. The ANSI codes are standard 24-bit RGB color codes: `\x1b[48;2;R;G;Bm` sets the background to RGB(R,G,B)."

**Likely questions**:

- Q: "Why did you choose Catmull-Rom interpolation?"
  A: "It's a good balance between speed and quality for downsampling. Nearest-neighbor would look blocky, and sinc interpolation would be slow. Catmull-Rom is built into golang.org/x/image and works well."

- Q: "What if the image fetch fails?"
  A: "We return an error string. The image display would show 'Error fetching image' instead of the album art. We could improve this by showing a placeholder or caching the image."

- Q: "Doesn't fetching a new image every time the track changes hit the network hard?"
  A: "Yes. In production, we should cache images. We could use a simple map with URL as key. That would be especially useful for albums that have multiple tracks."

---

## 5. Design Decisions

### 1. Single Monolithic Model Struct

**Decision**: All application state is stored in a single `Model` struct, not split into multiple smaller structures.

**Why**:

- Bubbletea expects this pattern. It makes the state flow predictable—all state changes go through `Update()`.
- Simpler mental model: one place to look for any piece of state.

**Alternatives**:

- Use multiple state structs with foreign keys (more complex, requires coordination)
- Use a database (overkill for a local TUI app)
- Use separate files with package-level variables (less testable, harder to reason about)

**Advantages**:

- Easy to debug: print the Model and see everything
- Avoids circular dependencies
- Bubbletea's tea.Model interface expects this pattern
- Easy to persist entire state to a file if needed

**Disadvantages**:

- Model struct is large and includes unrelated fields
- No type safety between different "screens" if app had multiple (it doesn't)
- Copying Model on every Update() is cheap but not ideal for large state

---

### 2. Generic HTTP Request Handler

**Decision**: Use Go 1.18 generics to write one `genericRequest[T]()` function for all HTTP operations.

**Why**:

- Eliminates boilerplate (would need 3 functions for GET/PUT/POST)
- Type-safe response parsing
- Easier to maintain (bugs fixed in one place)

**Alternatives**:

- Write separate functions for each method/response type
- Use interface{} and manual type assertion
- Use code generation

**Advantages**:

- DRY principle: all HTTP logic in one place
- Compiler verifies correct response types used
- Easy to add new endpoints

**Disadvantages**:

- Go generics add complexity (some developers unfamiliar with them)
- One-size-fits-all approach might not suit all future endpoints
- No specialization for specific endpoints if needed

---

### 3. Favorites Stored as Local JSON Files

**Decision**: User favorites are stored in `data/favorites/albums.json` and `data/favorites/playlists.json` instead of synced with Spotify.

**Why**:

- Spotify API doesn't have a "favorites" concept like Spotify's built-in favorites
- JukeTUI wanted a simple, independent favorites system
- Local files are simpler than a database

**Alternatives**:

- Use Spotify's built-in "Save Album" feature (but it's just all saved albums, no custom grouping)
- Use a database like SQLite
- Store favorites in a `~/.juketui/state.json` file

**Advantages**:

- Simple to implement (just read/write JSON)
- Human-readable data format
- No external dependencies
- Portable (follows user's home directory)

**Disadvantages**:

- Not synced across devices
- Manual file management required
- No backup/recovery if file corrupted
- Inefficient filtering logic during library fetches (favorites in memory as slice, O(n) lookups)

---

### 4. Polling Playback State Every 2 Seconds

**Decision**: Refresh playback state and progress bar on a 2-second timer instead of using Spotify's WebSocket or server-sent events.

**Why**:

- Spotify Web API doesn't offer WebSocket or event streams
- Simple to implement: just a timer and HTTP GET request
- Sufficient for a terminal UI where updates don't need to be real-time

**Alternatives**:

- Use Spotify's Web Playback SDK (requires browser, not suitable for TUI)
- Implement WebSocket to local Spotify desktop app (undocumented protocol)
- Poll every 1 second (uses more bandwidth)

**Advantages**:

- Simple to implement
- Stateless (each update is independent)
- Reasonable update frequency for UI

**Disadvantages**:

- Bandwidth inefficient (24-30 GET requests per minute)
- Not real-time (2-second delay for changes)
- If user pauses/unpauses externally, won't show until next poll
- Timer keeps running even if app is backgrounded

---

### 5. OAuth2 Callback Server on Localhost:8080

**Decision**: Spin up a local HTTP server to receive the OAuth2 authorization code instead of manually copying a callback URL.

**Why**:

- User experience: automatic redirect, no manual copy-paste
- Standard OAuth2 pattern for desktop applications

**Alternatives**:

- Use a redirect_uri that's an external server (requires infrastructure)
- Manual code copy-paste from browser
- Use device flow (requires user device registration)

**Advantages**:

- Seamless UX: user logs in, browser closes, app continues
- Standard pattern (many apps do this)
- Works offline after first auth

**Disadvantages**:

- HTTP server isn't cleaned up (keeps listening indefinitely)
- Could conflict if port 8080 is already in use
- Requires Spotify redirect_uri to be hardcoded as `http://127.0.0.1:8080/callback`
- Security: localhost is accessible to any process on the machine

---

### 6. Separate Tabs for Albums vs Playlists

**Decision**: Render two tabs (Albums, Playlists) with keyboard shortcuts (1, 2) to switch instead of a dropdown or mixed view.

**Why**:

- Clear visual separation of content types
- Keyboard navigation feels natural in a TUI
- Simple to implement (just re-fetch with different endpoint)

**Alternatives**:

- Mix albums and playlists in one list (harder to distinguish)
- Use a dropdown/modal selector
- Use a left sidebar with toggle

**Advantages**:

- Discoverable: tabs are visible on screen
- No modal/dropdown bloat
- Fast keyboard navigation

**Disadvantages**:

- Uses horizontal space
- Can't easily browse both at once
- Tab rendering is a bit fragile (hardcoded positions)

---

## 6. Technical Concepts Demonstrated

### 1. OAuth 2.0 Authorization Code Flow

**Demonstrated in**: [spotifyAuth.go](src/spotifyAuth.go)

- `OpenLoginPage()`: Opens authorization URL with client ID and requested scopes
- `GetCodeFromCallback()`: Receives authorization code from OAuth provider
- `GetSpotifyToken()`: Exchanges code for access token using client credentials

This is a real-world authentication pattern used by thousands of apps.

### 2. RESTful API Integration

**Demonstrated in**: [spotifyRequests.go](src/spotifyRequests.go), [spotifyHandlers.go](src/spotifyHandlers.go)

- `genericRequest()`: Low-level HTTP client for GET/PUT/POST
- Bearer token authentication in Authorization header
- Query parameters for pagination (limit, offset)
- JSON request/response marshaling

### 3. Event-Driven Architecture (Elm Architecture Pattern)

**Demonstrated in**: [main.go](src/main.go), Bubbletea framework

- Messages represent all state-changing events
- Commands represent async work
- Update() is the single state transition function
- View() is pure rendering from state

### 4. Pagination and Lazy Loading

**Demonstrated in**: [spotifyHandlers.go](src/spotifyHandlers.go#L59-L106)

- Offset-based pagination (limit=50, offset=0, offset=50, ...)
- Front-end offset management
- Page calculation in [uiUtil.go](src/uiUtil.go#L121-L132)

### 5. Asynchronous I/O Handling

**Demonstrated in**: [util.go](src/util.go), [main.go](src/main.go#L36-L45)

- `tea.Batch()` to run multiple concurrent commands
- `tea.Cmd` pattern for deferred work
- Goroutines spawned by commands don't block the UI

### 6. State Persistence (Local File I/O)

**Demonstrated in**: [json.go](src/json.go)

- `readJSON()`: Deserialize from file
- `writeJSONFile()`: Append to file
- `removeFromJSON()`: Filter and serialize
- Atomic writes using `os.WriteFile()`

### 7. Go Generics (Type Parameterization)

**Demonstrated in**: [spotifyRequests.go](src/spotifyRequests.go#L12-L58)

- `genericRequest[T any]()`: Generic type parameter for response
- `genericFetch[T]()`, `genericPut()`, `genericPost()`: Wrappers that specialize T
- Eliminates code duplication while maintaining type safety

### 8. Image Processing and Terminal Graphics

**Demonstrated in**: [uiUtil.go](src/uiUtil.go#L95-L165)

- Image decoding (PNG/JPG/GIF via `image` package)
- Image resizing with interpolation
- ANSI escape codes for 24-bit color
- Converting graphical data to terminal text

### 9. Dependency Injection via Environment Variables

**Demonstrated in**: [main.go](src/main.go#L318-L326), [util.go](src/util.go#L22-L35)

- `.env` file for configuration
- `queryEnv()` for defaults
- Keybinds configurable without code changes
- `SPOTIFY_ID`, `SPOTIFY_SECRET`, `SPOTIFY_PREFERENCE` externalized

### 10. Error Handling and Logging

**Demonstrated in**: [log.go](src/log.go), [spotifyHandlers.go](src/spotifyHandlers.go)

- Structured logging (error.log, data/logs/info.log)
- Conditional logging based on `DEVELOPMENT` environment variable
- Error logging at low level, error handling at high level
- Non-fatal errors logged but operation continues

### 11. Token Lifecycle and Refresh

**Demonstrated in**: [spotifyUtil.go](src/spotifyUtil.go#L27-L33), [spotifyAuth.go](src/spotifyAuth.go#L109-L130)

- Token expiry calculation (`expires_in` seconds)
- Proactive token refresh before expiration
- Refresh token flow for long-lived sessions

### 12. Cross-Platform Browser Integration

**Demonstrated in**: [spotifyAuth.go](src/spotifyAuth.go#L33-L44)

- Platform detection and browser launching:
  - `xdg-open` (Linux)
  - `open` (macOS)
  - `rundll32` (Windows)

---

## 7. Weaknesses and Improvements

### 1. **No Error Recovery or Offline Support**

**Issue**: If Spotify API is unreachable, UI shows stale data or error message. No retry logic.

**How it breaks**: Network blip during playback → playback state stops updating → UI looks frozen.

**How to discuss**: "Right now we log errors but keep showing old state. In a production app, I'd implement exponential backoff for retries. For offline scenarios, I could cache playback state and queue locally so basic info displays even if the API is down."

**Fix**: Add retry logic with exponential backoff, cache recent state in memory.

---

### 2. **Memory Leak: HTTP Server Never Shuts Down**

**Issue**: `GetCodeFromCallback()` starts an HTTP server that runs forever. When OAuth completes, the goroutine isn't cleaned up.

**Location**: [spotifyAuth.go](src/spotifyAuth.go#L58-L69)

**How it breaks**: Server is listening on :8080 even after successful login. If user re-runs app, port is still in use.

**How to discuss**: "There's a subtle bug here. The HTTP server that handles the OAuth callback keeps running even after we get the code. In a long-running TUI app this doesn't matter much, but if we ever want to log in again or restart, port 8080 might conflict. The fix is to create a proper HTTP server object and call `Shutdown()` once we get the callback."

**Fix**:

```go
server := &http.Server{Addr: ":8080"}
go func() {
  server.ListenAndServe()
}()
// ... wait for code ...
server.Shutdown(context.Background())
```

---

### 3. **Inefficient Favorites Filtering**

**Issue**: `handleFetchLibrary()` makes up to 2 API calls to filter out favorites. The filtering is O(n\*m) for each page.

**Location**: [spotifyHandlers.go](src/spotifyHandlers.go#L67-L92)

**How it breaks**: With 100 favorites and paginating through 1000 albums, we do expensive filtering repeatedly.

**How to discuss**: "The favorites filtering works but isn't efficient. We check every API result against every favorite using nested loops. If we have many favorites, this is O(n\*m). Also, if a favorite is on the first page, we re-fetch to skip it, which doubles the API calls. Better approach: precompute a set of favorite URIs for O(1) lookup. Or ask Spotify for unfavorited items only if the API supported it."

**Fix**:

```go
favoriteURIs := make(map[string]struct{})
for _, fav := range favorites {
  favoriteURIs[fav.URI] = struct{}{}
}
for _, item := range albums.Items {
  if _, isFavorite := favoriteURIs[item.Album.URI]; !isFavorite {
    filteredItems = append(filteredItems, item)
  }
}
```

---

### 4. **No Proper State Machine for App States**

**Issue**: `Model` includes all possible state fields even when not relevant. For example, `image` is only meaningful when a track is playing.

**Location**: [models.go](src/models.go)

**How it breaks**: No type-safety around valid state combinations. Could set `image` when `state.IsPlaying == false`.

**How to discuss**: "If I were designing this from scratch, I'd use a proper state machine or enum-based state. Right now the Model struct includes fields that might not make sense together. For example, if nothing is playing, what does `image` mean? In a larger app, I'd structure it as: `type AppState = NothingPlaying | Playing {currentTrack, image, progress} | LoggingIn {...}` using a sum type or interface."

**Fix**: Refactor to Rust-like enum state representation or tagged union in Go.

---

### 5. **No Caching for Album Art Images**

**Issue**: Every time a track changes, we fetch the album art image from Spotify's URL even if we've seen it before.

**Location**: [uiUtil.go](src/uiUtil.go#L160-L165)

**How it breaks**: For an album with 10 tracks, we fetch the same image 10 times.

**How to discuss**: "Each time we play a new track, we fetch its album art. For albums, all tracks have the same album art, so we're re-downloading the same image repeatedly. A simple fix is to cache images by URL. We could use a map[string]string of URL → rendered image, with a size limit to prevent memory explosion."

**Fix**: Add image cache in Model struct.

---

### 6. **Minimal Error Messages to User**

**Issue**: API errors are logged to `data/logs/errors.log` but user sees generic "Loading..." or stale data.

**Location**: [spotifyHandlers.go](src/spotifyHandlers.go#L17-L25), [main.go](src/main.go#L260-L265)

**How it breaks**: User doesn't know why something failed. No indication of what's wrong.

**How to discuss**: "Error handling is really basic right now. We log errors to a file but the user-facing UI just shows the last known state. In production, I'd add an error box to the UI that shows recent errors with timestamps. That way users know what went wrong instead of staring at a frozen display."

**Fix**: Display error messages in a dedicated UI area with timeout.

---

### 7. **No Tests**

**Issue**: No unit tests, integration tests, or even basic test coverage.

**Location**: No test files in repo

**How it breaks**: Refactoring is risky. Bugs go unnoticed.

**How to discuss**: "This is a personal project so I didn't prioritize tests, but in a production codebase I'd definitely add them. I'd at least have unit tests for the JSON parsing, favorites logic, and OAuth flow. Integration tests would mock the Spotify API. E2E tests would be tricky for a TUI but possible with pseudo-terminal automation."

**Fix**: Add `*_test.go` files, use `testify/assert` or similar.

---

### 8. **Token Refresh Not Guaranteed to Run**

**Issue**: `CheckTokenExpiryCmd()` is called but only if `Update()` is called. If app is backgrounded, timers stop, so refresh might not happen before token expires.

**Location**: [spotifyUtil.go](src/spotifyUtil.go#L27-L33), [main.go](src/main.go#L218)

**How it breaks**: If app sits idle for 50 minutes, token expires. Next interaction fails with 401 Unauthorized.

**How to discuss**: "Token refresh is tied to the Update() event loop. If the user backgrounds the app or doesn't interact with it for 50 minutes, the refresh timer doesn't fire and the token expires. The next interaction will fail. In production, I'd use a separate goroutine with a `time.Ticker` that runs independently and sends a message to the event loop when refresh is needed."

**Fix**: Add background goroutine for token refresh.

---

### 9. **Pagination Logic is Complex and Fragile**

**Issue**: Page calculation involves magic numbers (`UI_LIBRARY_SPACE`, `LIBRARY_SPACING`, height calculations). Wrapping around is complex.

**Location**: [main.go](src/main.go#L138-L170), [uiUtil.go](src/uiUtil.go#L121-L132)

**How it breaks**: Terminal resizing or changing `LIBRARY_SPACING` constant breaks pagination.

**How to discuss**: "The pagination logic works but it's a bit fragile. We have several magic numbers for spacing, and the logic for calculating pages is complex. It's not obvious why we subtract 3 in some cases but not others. This is code I'd refactor if continuing the project—extract a `PageCalculator` struct that encapsulates all the offset math."

**Fix**: Refactor pagination into a dedicated struct with methods.

---

### 10. **No Spotify Playback Verification**

**Issue**: We assume playback is active because `IsPlaying == true`. But if Spotify wasn't actually playing (device offline), we won't know.

**Location**: [main.go](src/main.go#L174-L187)

**How it breaks**: User selects a track to play, but Spotify isn't actually playing. UI shows "playing" anyway.

**How to discuss**: "We trust the `IsPlaying` flag from the playback state, but that doesn't mean audio is actually coming out of the speaker. If the Spotify desktop app crashes or the device goes offline, we won't detect it. Better approach: periodically verify that audio is playing by checking device status separately, or handle 400/500 responses from Spotify."

**Fix**: Add device connectivity check.

---

### 11. **No Handling for Expired/Revoked Tokens**

**Issue**: If user revokes app permissions in Spotify settings, next API call fails with 403. No recovery.

**Location**: [spotifyRequests.go](src/spotifyRequests.go#L51-L52)

**How it breaks**: App becomes non-functional, needs user to manually re-auth.

**How to discuss**: "If the user revokes the app's permissions in Spotify settings, we'll get a 403 Forbidden on the next API call. Right now we just log it and continue. Ideally we'd detect this and show a message asking the user to re-authorize, then restart the auth flow."

**Fix**: Handle 403 specifically and trigger re-auth flow.

---

### 12. **Single-Threaded Event Loop May Block on Slow Networks**

**Issue**: While a large API response is being downloaded or parsed, the UI doesn't respond to input.

**Location**: Async commands in util.go run in goroutines but parsing is in the goroutine, not the event loop

**How it breaks**: On a slow network, downloading a large playlist list might freeze the UI for a second.

**How to discuss**: "Commands run in goroutines so they shouldn't block the UI. But if the network is very slow or the JSON response is huge, there might be noticeable lag. In practice this isn't a big problem because Spotify API responses are usually small. But if we added features like full library search, large responses could cause visible jank."

**Fix**: Use streaming JSON parsing, or show progress indicator during large fetches.

---

### 13. **UX: No Visual Feedback for Favorites Toggle**

**Issue**: When user presses 'f' to favorite/unfavorite, there's a re-fetch and the list refreshes, but it's not obvious what happened.

**Location**: [main.go](src/main.go#L86-L120)

**How it breaks**: User presses 'f', list briefly flickers while loading, user doesn't know if their action worked.

**How to discuss**: "Toggling favorites works, but the UX is a bit jarring. The list disappears, refetches, and re-renders. I'd add visual feedback like temporarily highlighting the item with a checkmark, or showing a brief toast message."

**Fix**: Add success message or highlight animation.

---

### 14. **No Handling for Empty Library**

**Issue**: If user has no albums or playlists, app might show confusing UI.

**Location**: [uiUtil.go](src/uiUtil.go#L121-L132)

**How it breaks**: Pagination math breaks with zero items. No special case.

**How to discuss**: "If the user has no saved albums, the pagination calculation will be weird. We should detect empty library and show a message like 'No albums found. Go favorite some albums in Spotify!'"

**Fix**: Add empty state UI.

---

### 15. **Security: Redirect URI Hardcoded**

**Issue**: OAuth redirect URI is hardcoded as `http://127.0.0.1:8080/callback`. It's HTTP, not HTTPS.

**Location**: [spotifyAuth.go](src/spotifyAuth.go#L23)

**How it breaks**: For a personal project, this is fine. But in production on a shared machine, other processes could intercept the auth code.

**How to discuss**: "For a local TUI app, HTTP redirect URI is acceptable because it's only accessible locally. But ideally we'd use a dynamic port number to avoid conflicts, or at minimum check that the redirect URL matches what we expect in Spotify settings."

**Fix**: Make port configurable, use random available port.

---

## 8. Interview Questions (20+) with Answers

### Architecture & Design

**Q1: Walk me through the architecture of JukeTUI. How do the pieces fit together?**
A: "JukeTUI uses Bubbletea's Elm-inspired architecture. At the center is the Model struct that holds all state. The event loop in Update() processes all messages—keyboard input, API responses, timer ticks. Based on the message, we update Model and return Commands for async work. Commands run in goroutines and send messages back to Update(). View() just renders the current Model. For Spotify integration, we have a generic HTTP request handler that abstracts all API communication. Authentication uses OAuth2 with a local callback server."

---

**Q2: Why did you choose to use Bubbletea for the UI instead of building with a different framework?**
A: "Bubbletea is the most popular TUI framework in Go. It's well-maintained and handles a lot of complexity for terminal rendering and event handling. The Elm architecture (Update/View/Cmd) is clean and makes state management predictable. Alternatives would be ncurses bindings (low-level, more boilerplate) or building raw terminal control (very error-prone)."

---

**Q3: How does the favorites system work?**
A: "Favorites are stored as JSON in `data/favorites/albums.json` and `data/favorites/playlists.json`. When we fetch the library, we filter out favorites from the API results and prepend them to the top of the list. The filtering is done in-memory after fetching. This gives users a way to pin albums/playlists without using Spotify's built-in save feature."

---

**Q4: The pagination logic looks complex. How does it work?**
A: "We use Spotify's offset-based pagination API (limit=50, offset=0, offset=50, etc.). On the frontend, we calculate which items to show based on terminal height. When the user presses right arrow (next page), we increment the offset and re-fetch. The trickiest part is accounting for favorites—they reduce the number of items we need to fetch to fill the page."

---

### Authentication & Security

**Q5: Explain your OAuth2 implementation. How does the login flow work?**
A: "We use OAuth2's authorization code grant type. First, we construct an auth URL with our client ID and requested scopes. We open it in the user's browser. They log in and grant permissions. Spotify redirects to our local callback server on :8080 with an authorization code. We extract the code from the URL, then POST it to Spotify's token endpoint along with our client secret (in Basic auth header). Spotify returns an access token (valid 1 hour) and refresh token (long-lived). We store both and use the access token for all subsequent API calls."

---

**Q6: How do you handle token expiration?**
A: "We track when the token expires (current time + expires_in seconds). Before each Update() cycle, we check if we're past the expiration time. If so, we call RefreshSpotifyToken() which POSTs to Spotify's token endpoint with the refresh token. Spotify returns a new access token with a new expiration time. Subsequent API calls use the new token. The refresh is transparent to the user."

---

**Q7: What if the user revokes permissions or the token becomes invalid?**
A: "Right now, we'd get a 403 Forbidden from Spotify API. We log the error but don't have specific recovery logic. Ideally, we'd detect this scenario and guide the user to re-authenticate. That's a limitation in the current implementation."

---

**Q8: Why did you put the redirect URI on localhost instead of an external server?**
A: "For a desktop app, localhost is the standard pattern. It's simple—we don't need infrastructure. The tradeoff is that port :8080 could be in use or could require port forwarding if the app is run in a container. For a production app, I'd make the port configurable or dynamically choose an available port."

---

### API Integration

**Q9: How is the Spotify API integrated? Walk me through the request path.**
A: "All Spotify API calls go through `genericRequest[T]()` in spotifyRequests.go. This function builds an HTTP request to the Spotify API (base URL: `https://api.spotify.com/v1`). We add the Authorization header with the Bearer token. The caller specifies the response type T (e.g., PlaybackState), and for GET requests we parse the JSON into that type. For PUT/POST we just return the status code. Higher-level functions like `handleFetchPlayback()` wrap this and return Bubbletea commands."

---

**Q10: You're using Go generics for the HTTP client. Why not use interface{}?**
A: "Generics give us compile-time type safety. With interface{}, the type is lost at runtime—you could accidentally assign a PlaybackState to a variable expecting SpotifyAlbum. With generics, the compiler verifies the type at call sites. It's cleaner and prevents whole classes of bugs."

---

**Q11: What endpoints does the app use, and what does each one do?**
A: "

- `GET /me/player` → Get current playback state (track, position, device, shuffle status)
- `PUT /me/player/play` → Start playback with optional context_uri
- `PUT /me/player/pause` → Pause playback
- `POST /me/player/next` → Skip to next track
- `PUT /me/player/shuffle` → Toggle shuffle
- `GET /me/albums` → Get user's saved albums (paginated)
- `GET /me/playlists` → Get user's playlists (paginated)
- `GET /me/player/queue` → Get the queue (next ~20 tracks)

All of these require appropriate OAuth scopes."

---

**Q12: How does the app handle Spotify API errors or rate limits?**
A: "Poorly, if I'm honest. We check for 403 (permissions) specifically. Other errors get logged but we return an empty struct. There's no retry logic, exponential backoff, or rate limit handling. In production, I'd implement retries with backoff, and maybe a queue of pending requests to respect rate limits."

---

### Async & Concurrency

**Q13: How are asynchronous operations handled in Bubbletea?**
A: "Bubbletea uses a command pattern. When we need to do I/O (fetch API, process image, etc.), we return a `tea.Cmd` which is a function that runs in a goroutine. When it completes, it sends a message back to the Update() loop. All state updates happen in the single-threaded Update(), so there's no need for mutexes. Multiple commands can run concurrently (via `tea.Batch()`), but their results are serialized back to the event loop."

---

**Q14: There's a polling loop that fetches playback state every 2 seconds. Why polling instead of webhooks or server-sent events?**
A: "Spotify's Web API doesn't offer webhooks or WebSocket connections. Polling is the only option with the REST API. Two seconds is a reasonable trade-off between responsiveness and bandwidth. For a TUI where updates don't need to be instant, it works fine."

---

**Q15: What happens if an API call is slow? Does the UI freeze?**
A: "The API call runs in a goroutine spawned by a command, so the main event loop isn't blocked. However, if the network is very slow or the response is huge, there could be a brief freeze when parsing the JSON in the goroutine or when rendering a large response. In practice, Spotify's responses are small so this rarely happens."

---

### Image Processing

**Q16: How do you render album art in a terminal? That seems hard.**
A: "It's pretty creative! We fetch the album art image from Spotify's URL, resize it to fit the terminal dimensions using high-quality Catmull-Rom interpolation, then convert each pixel to an ANSI escape code that sets the terminal background color. Each pixel becomes roughly a 2-character block with that color. The result is a small pixelated image that's still recognizable."

---

**Q17: Do you cache album images to avoid re-fetching?**
A: "No, we don't cache them currently. Every time the track changes, we fetch the image from Spotify again. For albums with multiple tracks, this means downloading the same image multiple times. It's inefficient. A simple improvement would be to cache images by URL in a map."

---

### State & Data

**Q18: How do you persist user state? Where is data stored?**
A: "The only persistent data is favorites, which are stored as JSON arrays in `data/favorites/albums.json` and `data/favorites/playlists.json`. Everything else (playback state, queue, etc.) is fetched from Spotify on each update. The app doesn't save session state, so if you restart it, it starts fresh with the current playback from Spotify."

---

**Q19: The Model struct is pretty large. How do you manage it?**
A: "All state lives in one Model struct. It's not ideal for a large app, but for JukeTUI's scope it's fine. Every state change goes through Update(), which makes the flow predictable. If the app grew much larger, I'd refactor into a proper state machine with separate types for different app states (e.g., LoginState, PlayingState, BrowsingState)."

---

### Testing & Quality

**Q20: Why are there no tests in the project?**
A: "This is a personal project I built for fun, and testing wasn't a priority. But if I were continuing it or deploying it to users, I'd definitely add tests. At minimum: unit tests for JSON parsing and favorites logic, integration tests with a mocked Spotify API, and possibly e2e tests with a fake Spotify server."

---

**Q21: If you were adding tests now, where would you start?**
A: "I'd start with the JSON parsing and favorites logic in json.go and util.go—they're pure functions and easy to test. Then mock out the Spotify API and test the pagination and filtering logic. Finally, I'd test the OAuth flow by mocking the HTTP server."

---

**Q22: What logging do you have?**
A: "We have two loggers (errorLogger and infoLogger) that write to error.log and data/logs/info.log respectively. They only log if DEVELOPMENT=true in .env. Errors are logged at the point of failure, and info logs successful API requests. It's very basic—there's no log level filtering or structured logging. For production, I'd use a proper logging library."

---

### Future & Improvements

**Q23: If you were continuing this project today, what would you change first?**
A: "Top three things: (1) Fix the HTTP server cleanup—it leaks goroutines. (2) Improve error handling—show errors to the user with proper recovery flows. (3) Add tests. Beyond that, I'd refactor the pagination and favorites logic which is fragile, improve the favorites filtering to use a set for O(1) lookups, and add image caching."

---

**Q24: How would you handle offline mode or poor network connectivity?**
A: "I'd implement retry logic with exponential backoff. For offline scenarios, I'd cache critical state (playback position, queue) so basic info displays even if the API is down. I'd also show the user explicit status: 'Offline mode', 'Retrying...', etc. instead of silently failing."

---

**Q25: How might you scale this if it grew into a multi-user or collaborative application?**
A: "As-is, it's single-user and local. To scale: (1) Move state to a backend database and sync over the network. (2) Add multi-user support with user accounts and permissions. (3) Replace local JSON with server-based storage. (4) Use WebSocket or gRPC instead of REST for real-time updates. The architecture would shift from 'single Bubbletea loop' to 'TUI client + backend API'."

---

## 9. Live Walkthrough Plan (10-15 Minutes)

**Goal**: Demonstrate the app, explain key architecture decisions, and showcase interesting code.

### Timeline & Talking Points

**0:00-1:00 - Introduction & Demo (1 minute)**

- "JukeTUI is a terminal UI for controlling Spotify. Let me show you what it looks like and does."
- Run the app: `go run .`
- Point out the layout: library on the left (albums/playlists with pagination), album art in the middle (pixelated), queue at the bottom, playback controls on top.
- Show the keybinds: press 'p' to pause, 'n' to skip, '1' to switch to albums, '2' to switch to playlists.
- Demonstrate navigating with arrow keys, changing pages with left/right arrows.

**1:00-2:30 - OAuth Flow & Authentication (1.5 minutes)**

- "Behind the scenes, this uses OAuth2 to log in to Spotify."
- Open [spotifyAuth.go](src/spotifyAuth.go)
- Explain the flow: we build an auth URL, open it in the browser, capture the code from the redirect, exchange it for an access token.
- Point out `GetSpotifyToken()` and `RefreshSpotifyToken()`.
- Key point: "The token lasts 1 hour, so we use the refresh token to get a new one automatically. The user never has to log in again."

**2:30-4:00 - API Communication Layer (1.5 minutes)**

- Open [spotifyRequests.go](src/spotifyRequests.go) and [spotifyHandlers.go](src/spotifyHandlers.go)
- Show `genericRequest[T]()`: "This is a generic function that handles GET/PUT/POST to Spotify's API. It builds the request, adds the Bearer token, and parses the JSON response into any type T we specify."
- Show `handleFetchPlayback()` and `handleFetchLibrary()`: "These wrap the low-level HTTP code with Bubbletea's command pattern. The actual work runs in a goroutine and sends messages back to the UI loop."
- Key point: "Using generics, we avoid code duplication. Spotify has 10+ endpoints but we use one generic function."

**4:00-6:00 - State Management & Event Loop (2 minutes)**

- Open [main.go](src/main.go) and [models.go](src/models.go)
- Show Model struct: "All app state lives here. The event loop in Update() processes every message—keyboard input, API responses, timers."
- Show a simple user interaction: press 'p' (play/pause)
  - Trace through the code: `Update(tea.KeyMsg)` → detects 'p' → calls `handleGenericPut("/me/player/pause", ...)` → returns a command
  - Point out: "The command runs async, doesn't block the UI. The message comes back, Update() runs again, View() renders."
- Key point: "This Elm-inspired pattern makes state flows predictable."

**6:00-8:00 - Image Processing (2 minutes)**

- Open [uiUtil.go](src/uiUtil.go)
- Show `makeNewImage()` and `printImage()`.
- Explain: "We download the album art image from Spotify, resize it using interpolation, then convert each pixel to an ANSI escape code for the terminal. Each pixel becomes a colored 2-character block."
- Key point: "This is the 'retro' feel JukeTUI has. Most terminal apps can't display images, so we convert them to colored text."

**8:00-9:30 - Favorites & Pagination (1.5 minutes)**

- Show [json.go](src/json.go) and favorites logic in [main.go](src/main.go)
- Explain: "Favorites are stored locally in JSON. When browsing the library, we fetch from Spotify and filter out favorites. Favorites always appear at the top."
- Explain pagination: "Spotify returns items in batches of 50. We manage offset and limit to show multiple pages."
- Key point: "This is a bit complex because we need to coordinate between local favorites and API pagination. It's something I'd refactor if continuing the project."

**9:30-10:30 - Async & Concurrency (1 minute)**

- Point to `Init()` in [main.go](src/main.go): "See how we use `tea.Batch()` to run multiple commands at startup? All of these run concurrently in goroutines. But all state updates happen in a single-threaded event loop. That's the key insight: work is parallel but state updates are serialized."
- Show `scheduleNextFetch()` in [util.go](src/util.go): "We poll Spotify every 2 seconds for playback updates. This command sleeps and then sends a message back to the loop."

**10:30-12:00 - Design Decisions & Tradeoffs (1.5 minutes)**

- "Here are some design decisions I'm proud of and some I'd change:"
- **Good**: Generic HTTP client, Bubbletea pattern (predictable state flow), OAuth token refresh
- **Could improve**: (1) The HTTP server from OAuth doesn't shut down (memory leak), (2) Favorites filtering is inefficient, (3) No tests, (4) Error handling is minimal
- Key point: "Personal projects are great for experimenting, but you learn what to do differently."

**12:00-13:30 - Questions (1.5 minutes)**

- "I'll stop here and take questions. Some things we could dive deeper on: image processing, the OAuth flow, or architectural choices. Any thoughts?"

---

## 10. Cheat Sheet

### 30-Second Project Pitch

"JukeTUI is a terminal user interface for controlling Spotify written in Go. It uses OAuth2 to authenticate, fetches your album library and playlists from Spotify's API, displays album art as colored characters in the terminal, and lets you control playback with keyboard shortcuts. It demonstrates async I/O patterns, API integration, and the Elm-inspired state architecture."

---

### Architecture Summary

- **Language**: Go 1.23.2
- **Framework**: Bubbletea (TUI) + Lipgloss (styling)
- **Pattern**: Elm architecture (Update/View/Cmd)
- **Auth**: OAuth2 authorization code grant
- **API**: Spotify Web API (REST + Bearer tokens)
- **State**: Single Model struct, mutations in Update()
- **Persistence**: Local JSON files for favorites
- **Rendering**: ANSI escape codes for colored terminal output

---

### 5 Most Important Files

1. **[main.go](src/main.go)** - Entry point, event loop (Update/View), initialization
2. **[spotifyAuth.go](src/spotifyAuth.go)** - OAuth login and token refresh
3. **[spotifyHandlers.go](src/spotifyHandlers.go)** - High-level API operations (fetch playback, library, queue)
4. **[uiUtil.go](src/uiUtil.go)** - Rendering (library list, playback bar, album art)
5. **[json.go](src/json.go)** - Persistence (favorites read/write)

---

### 5 Most Important Technical Decisions

1. **Elm Architecture (Bubbletea)**: Predictable state flow, all mutations through Update()
2. **Go Generics for HTTP**: Type-safe API client, no code duplication
3. **Local JSON Favorites**: Simple persistence independent of Spotify
4. **OAuth2 Local Callback**: Seamless auth UX without external infrastructure
5. **Async Commands**: Concurrent I/O without blocking the UI

---

### 5 Key Technical Accomplishments

1. **OAuth2 Implementation**: Complete authentication flow with token refresh
2. **Image-to-Terminal Conversion**: Pixel-based rendering of album art
3. **Pagination & Filtering**: Coordinate between API pagination and local favorites
4. **Generic HTTP Client**: Reusable, type-safe API communication layer
5. **Responsive TUI**: Non-blocking async I/O using Elm-inspired command pattern

---

### 5 Weaknesses/Improvements

1. **HTTP Server Cleanup**: OAuth callback server never shuts down
2. **Favorites Filtering**: O(n\*m) inefficiency, possible double API calls
3. **No Tests**: Zero test coverage
4. **Error Handling**: Errors logged but not shown to user
5. **No Image Caching**: Album art re-fetched for every track change

---

### 10 Likely Interview Questions (TL;DR Answers)

1. **"What does JukeTUI do?"** → Terminal UI for Spotify playback control
2. **"Explain the OAuth flow."** → Authorization code grant: open browser → capture code → exchange for token
3. **"How is state managed?"** → Single Model struct, all mutations in Update()
4. **"Why Go generics for HTTP?"** → Type-safe, eliminates duplication
5. **"How do you render images in terminal?"** → Convert pixels to ANSI background colors
6. **"How is async handled?"** → Commands return goroutines, results become messages
7. **"How does pagination work?"** → Offset-based, accounts for local favorites
8. **"What's the architecture pattern?"** → Elm-inspired: Update/View/Cmd
9. **"How are favorites persisted?"** → Local JSON files, filtered from API results
10. **"What would you change today?"** → Fix server cleanup, improve error handling, add tests

---

---

## Appendix: Code Quality Notes

### Strengths

- Clean separation of concerns (auth, API, UI, persistence in separate files)
- Good use of Go features (generics, type-safe error handling)
- Bubbletea pattern makes state flows obvious
- Well-commented code with clear function purposes
- Proper use of OAuth2 standard

### Weaknesses

- No tests or error recovery
- Some magic numbers (LIBRARY_SPACING, UI_LIBRARY_SPACE)
- Pagination math is complex and fragile
- Minimal logging and diagnostics
- HTTP server from OAuth never shuts down

### If Continuing the Project

- **Priority 1**: Fix HTTP server cleanup and error handling
- **Priority 2**: Add tests and proper logging
- **Priority 3**: Refactor pagination and favorites logic
- **Priority 4**: Add image caching
- **Priority 5**: Improve UX feedback for operations

---

**Document generated for interview preparation. Questions to ask your interviewer if they ask follow-ups: "Do you want me to dive deeper into any of these areas?" or "Would you like to see specific code samples?"**
