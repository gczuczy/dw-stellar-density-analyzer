DROP SCHEMA IF EXISTS density CASCADE;

CREATE SCHEMA density;
GRANT USAGE ON SCHEMA density TO edadmin, edservice, edviewer;

CREATE TABLE density.projects (
       id int GENERATED ALWAYS AS IDENTITY,
       name varchar(64) NOT NULL,
       PRIMARY KEY (id)
);
GRANT SELECT, INSERT, UPDATE, DELETE ON density.projects TO edservice;
GRANT SELECT ON density.projects TO edviewer;
INSERT INTO density.projects (name) VALUES
('A15X CW Density Scans'),
('DW3 Stellar Density Scans'),
('DW3 Logarithmic Density Scans')
;

CREATE TABLE density.cmdrs (
       id    int		  GENERATED ALWAYS AS IDENTITY,
       name  varchar(64)	  NOT NULL UNIQUE,
       PRIMARY KEY (id)
);
GRANT SELECT, INSERT, UPDATE ON density.cmdrs TO edservice;
GRANT SELECT ON density.cmdrs TO edviewer;

CREATE TABLE density.measurements (
       id    int		  GENERATED ALWAYS AS IDENTITY,
       projectid int		  NOT NULL,
       cmdrid	 int		  NOT NULL,
       FOREIGN KEY (projectid) REFERENCES density.projects (id),
       FOREIGN KEY (cmdrid) REFERENCES density.cmdrs(id),
       PRIMARY KEY (id)
);
GRANT SELECT, INSERT ON density.measurements TO edservice;
GRANT SELECT ON density.measurements TO edviewer;

CREATE TABLE density.datapoints (
       id    int		  GENERATED ALWAYS AS IDENTITY,
       measurementid int	  NOT NULL,
       sysname	     varchar(64)  NOT NULL,
       zsample	     int	  NOT NULL,
       x	     real	  NOT NULL,
       y	     real	  NOT NULL,
       z	     real	  NOT NULL,
       syscount	     int	  NOT NULL,
       maxdistance   real	  NOT NULL,
       FOREIGN KEY (measurementid) REFERENCES density.measurements(id),
       PRIMARY KEY (id),
       UNIQUE (measurementid, zsample),
       UNIQUE (measurementid, sysname),
       CHECK (syscount >= 0 AND syscount <= 50),
       CHECK (maxdistance > 0 AND maxdistance <= 20)
);
GRANT SELECT, INSERT ON density.datapoints TO edservice;
GRANT SELECT ON density.datapoints TO edviewer;
