CREATE TABLE users
(
    id           int          NOT NULL AUTO_INCREMENT,
    last_ip      varchar(15)  NOT NULL,
    display_name varchar(255) NOT NULL,
    role_id      int          NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (role_id) REFERENCES roles(id)
);