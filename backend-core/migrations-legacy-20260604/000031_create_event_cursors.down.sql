-- migration 000031 down: drop event_cursors table.

DROP TABLE IF EXISTS event_cursors;
