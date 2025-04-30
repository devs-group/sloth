-- +goose Up
CREATE TABLE persistent_volumes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    volume_name TEXT NOT NULL,
    mount_path TEXT NOT NULL,

    -- Foreign Key
    -- Can be null because the service doesn't exist yet
    service_id INTEGER,
    -- In case of a migration from an older volume
    source_volume_id INTEGER,

    CONSTRAINT FK_PersistentVolume_Service FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
    CONSTRAINT FK_PersistentVolume_PersistentVolume FOREIGN KEY (source_volume_id) REFERENCES persistent_volumes(id) ON DELETE SET NULL,

    CONSTRAINT UQ_VolumeName UNIQUE(volume_name)
);

-- +goose Down
DROP TABLE IF EXISTS persistent_volumes;