DELETE FROM features WHERE key IN (
    'import-profile',
    'dispatch-import-profile',
    'import-autonomera-profiles',
    'dispatch-import-autonomera-profiles'
);

ALTER TABLE offers DROP COLUMN IF EXISTS profile_external_id;

DROP TABLE IF EXISTS profiles;

DROP TABLE IF EXISTS users;
