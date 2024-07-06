This is a project that lists events related to AI horde.

All events are written into [source.json](source.json) and must follow the schema defined in [schema.json](schema.json).

On every push to the main branch, the `source.json` file is validated against `schema.json` and if it passes 
(including custom checks, like making sure each event's ID is truly unique), various optimized versions of the file
are generated.

Those optimized versions are stripped of events that are not yet valid, or are no longer valid, or events that are not
related to your project (more on that below).

Each event can be constrained to one or more projects, any string key is valid. Such constrained events are excluded
from other project-specific results and from the common results.

If you want to ensure your project-specific JSON gets generated even though you don't have any project-specific events
yet, just add it to [clients.json](clients.json).

Currently generated files:

- [results.common.json](results.common.json) and [results.common.min.json](results.common.min.json) - these two are usable
  by any project and don't contain any project-specific events
- [results.horde-ng.json](results.horde-ng.json) and [results.horde-ng.min.json](results.horde-ng.min.json) - these
  contain both the common events and events related only to HordeNG
- [results.artbot.json](results.artbot.json) and [results.artbot.min.json](results.artbot.min.json) - these
  contain both the common events and events related only to Artbot
