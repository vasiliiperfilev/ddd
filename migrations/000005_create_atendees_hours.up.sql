CREATE TABLE atendees_hours
(
    id           int                                                       NOT NULL AUTO_INCREMENT,
    hour_id      int                                                       NOT NULL,
    atendee_id   int                                                       NOT NULL,
    PRIMARY KEY (id)
);