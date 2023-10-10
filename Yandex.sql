CREATE DATABASE ycloudlogstorage;
\c ycloudlogstorage;
CREATE TABLE vmInfo (
                        dbid SERIAL PRIMARY KEY,
                        machineId VARCHAR(32) UNIQUE,
                        nameMachine TEXT,
                        description TEXT,
                        folderId VARCHAR(32),
                        bootDiskId VARCHAR(32),
                        ram BIGINT,
                        coresCount smallint,
                        zoneId VARCHAR(32),
                        platformId VARCHAR(32),
                        serviceAccountId VARCHAR(32)
);
CREATE TABLE disksInfo (
                           dbid SERIAL PRIMARY KEY,
                           diskId VARCHAR(32) UNIQUE,
                           folderId VARCHAR(32),
                           typeId VARCHAR(32),
                           zoneId VARCHAR(32),
                           sizeDisk BIGINT
);
CREATE TABLE serviceAccsInfo (
                                 dbid SERIAL PRIMARY KEY,
                                 nameSAcc TEXT,
                                 description TEXT,
                                 serviceAccId VARCHAR(32) UNIQUE,
                                 folderId VARCHAR(32)
);
CREATE TABLE vmAvgLoadHost (
                               dbid BIGSERIAL PRIMARY KEY,
                               cpu INT,
                               ram BIGINT
);
CREATE TABLE editLog (
                         dbid BIGSERIAL PRIMARY KEY,
                         columnEdit VARCHAR(32),
                         oldDataColumn VARCHAR(32),
                         newDataColumn VARCHAR(32),
                         timeEdited BIGINT
);
