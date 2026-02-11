# KANM Now Playing
SSE microservice that broadcasts what's currently playing on the radio.
## How it works
POST a track update → all connected listeners get it instantly via Server-Sent Events.
## Usage
### Listen to updates
```bash
curl -N http://localhost:8000/now-playing
```
```javascript
const eventSource = new EventSource('http://localhost:8000/now-playing');
eventSource.onmessage = (event) => {
  const track = JSON.parse(event.data);
  console.log(`${track.songName} by ${track.artistName}`);
};
```
You get the current track immediately on connect, then updates as they happen.
### Update the track

**curl:**
```bash
curl -X POST http://localhost:8000/update \
  -H "X-API-Key: your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{
    "songName": "Wonderwall",
    "artistName": "Oasis",
    "albumName": "(Whats the Story) Morning Glory?"
  }'
```

**fetch (Next.js):**
```javascript
await fetch('http://localhost:8000/update', {
  method: 'POST',
  headers: {
    'X-API-Key': process.env.SYSTEM_PSK,
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    songName: 'Wonderwall',
    artistName: 'Oasis',
    albumName: "(What's the Story) Morning Glory?",
  }),
  cache: 'no-store',
});
```

**axios:**
```javascript
await axios.post('http://localhost:8000/update', {
  songName: 'Wonderwall',
  artistName: 'Oasis',
  albumName: "(What's the Story) Morning Glory?",
}, {
  headers: {
    'X-API-Key': process.env.SYSTEM_PSK,
  },
});
```

This broadcasts to all connected listeners.
## Deployment
Set `SYSTEM_PSK` environment variable (used for the `X-API-Key` header):
```bash
docker build -t kanm-now-playing .
docker run -p 8000:8000 -e SYSTEM_PSK=your-secret-here kanm-now-playing
```
Or use a `.env` file:
```bash
docker run -p 8000:8000 --env-file .env kanm-now-playing
```
## TODO (for the user/integrator with this microservice)
Integrate with Next.js app to auto-update from Bitjockey.

```
WHEN song changes
 IF user is in station
  POST <kanm-now-playing-url>/update with {...}
```
