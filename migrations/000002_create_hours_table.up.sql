CREATE TABLE 'hours'
(
    hour         TIMESTAMP                                                 PRIMARY KEY,
    availability ENUM ('available', 'not_available', 'training_scheduled') NOT NULL,
);