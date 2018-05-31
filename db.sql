CREATE TABLE departments (
        id INT NOT NULL DEFAULT unique_rowid(),
        "name" STRING NULL,
        FAMILY "primary" (id, "name", rowid)
);

CREATE TABLE memorandums (
        id INT NOT NULL DEFAULT unique_rowid(),
        accepted INT NULL DEFAULT 0:::INT,
        addtime DATE NULL DEFAULT '2017-08-07':::DATE,
        departmentid INT NULL,
        CONSTRAINT "primary" PRIMARY KEY (id ASC),
        FAMILY "primary" (id, accepted, addtime, departmentid)
);

CREATE TABLE wifiusers (
        id INT NOT NULL DEFAULT unique_rowid(),
        mac STRING NOT NULL,
        username STRING NULL,
        phonenumber STRING NULL,
        hash STRING NULL,
        memorandumid INT NULL,
        accepted INT NULL DEFAULT 0:::INT,
        disabled INT NULL DEFAULT 0:::INT,
        departmentid INT NULL,
        CONSTRAINT "primary" PRIMARY KEY (id ASC),
        FAMILY "primary" (id, mac, username, phonenumber, hash, memorandumid, accepted, disabled, departmentid)
);

CREATE TABLE ethmemorandums (
        id INT NOT NULL DEFAULT unique_rowid(),
        accepted INT NULL DEFAULT 0:::INT,
        addtime DATE NULL DEFAULT '2017-08-07':::DATE,
        department STRING NULL,
        hash STRING NULL,
        CONSTRAINT "primary" PRIMARY KEY (id ASC),
        FAMILY "primary" (id, accepted, addtime, department, hash)
);

CREATE TABLE ethusers (
        id INT NOT NULL DEFAULT unique_rowid(),
        mac STRING NOT NULL,
        class STRING NULL,
        building STRING NULL,
        info STRING NULL,
        memorandumid INT NULL,
        CONSTRAINT "primary" PRIMARY KEY (id ASC),
        FAMILY "primary" (id, mac, class, building, info, hash, memorandumId, accepted)
);


CREATE TABLE phonememorandums (
        id INT NOT NULL DEFAULT unique_rowid(),
        accepted INT NULL DEFAULT 0:::INT,
        addtime DATE NULL DEFAULT '2017-08-07':::DATE,
        exist INT NULL,
        department STRING NULL,
        hash STRING NULL,
        CONSTRAINT "primary" PRIMARY KEY (id ASC),
        FAMILY "primary" (id, accepted, addtime, department, hash)
);


CREATE TABLE phoneusers (
        id INT NOT NULL DEFAULT unique_rowid(),
        info STRING NOT NULL,
        access INT NULL,
        memorandumid INT NULL,
        phone STRING NULL,
        CONSTRAINT "primary" PRIMARY KEY (id ASC),
        FAMILY "primary" (id, info, access, memorandumid, phone)
);

CREATE TABLE exphoneusers (
        id INT NOT NULL DEFAULT unique_rowid(),
        info STRING NOT NULL,
        class STRING NULL,
        building STRING NULL,
        memorandumid INT NULL,
        phone STRING NULL,
        CONSTRAINT "primary" PRIMARY KEY (id ASC),
        FAMILY "primary" (id, info, access, memorandumid, phone)
);


CREATE TABLE mailmemorandums (
        id INT NOT NULL DEFAULT unique_rowid(),
        accepted INT NULL DEFAULT 0:::INT,
        addtime DATE NULL DEFAULT '2017-08-07':::DATE,
        department STRING NULL,
        reason STRING NULL,
        hash STRING NULL,
        CONSTRAINT "primary" PRIMARY KEY (id ASC),
        FAMILY "primary" (id, accepted, addtime, department, hash, reason)
);

CREATE TABLE mailusers (
        id INT NOT NULL DEFAULT unique_rowid(),
        mail STRING NOT NULL,
        name STRING NULL,
        position STRING NULL,
        memorandumid INT NULL,
        CONSTRAINT "primary" PRIMARY KEY (id ASC),
        FAMILY "primary" (id, mail, name, position, memorandumid, hash, memorandumId, accepted)
);