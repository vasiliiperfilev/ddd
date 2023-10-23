CREATE TABLE trainers_hours
(
    id           int                                                       NOT NULL AUTO_INCREMENT,
    hour_id      int                                                       NOT NULL,
    trainer_id   int                                                       NOT NULL,
    availability ENUM ('available', 'not_available', 'training_scheduled') NOT NULL,
    PRIMARY KEY (id)
);