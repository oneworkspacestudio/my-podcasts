# Podcast hosts and private intro tracks

Preview.7 changes both writing/editorial routes so Alex and Jamie stay in
character. The app adds each name to its own spoken turn, and episode pages
and transcript exports label AI narration. Existing scripts and ready audio
retain their content.

The account can upload two MP3/WAV tracks under Create. The browser converts
them to bounded PCM. SQLite stores the private pair and pins one random track
when the episode is queued. Mastering puts it after the first spoken chapter
and recalculates the remaining chapter offsets. The migration is additive.
User-supplied tracks stay out of public source and release assets.

My Voice preview.10 supplies the male Marius and female Alba stock presets on
Pocket TTS. The compatible cast field pins the mapping and audio cache.
Reinstall an older Pocket model to fetch its second stock preset. An existing
single-preset engine can continue to render. Keep the 48-payload namespace.
