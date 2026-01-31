DROP SCHEMA IF EXISTS density CASCADE;

CREATE SCHEMA density;
GRANT USAGE ON SCHEMA density TO edadmin, edservice, edviewer;

CREATE TABLE density.campaigns (
       id int GENERATED ALWAYS AS IDENTITY,
       name varchar(64) NOT NULL,
       PRIMARY KEY (id)
);
GRANT SELECT, INSERT, UPDATE, DELETE ON density.campaigns TO edservice;
GRANT SELECT ON density.campaigns TO edviewer;
INSERT INTO density.campaigns (name) VALUES
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

CREATE TABLE density.surveys (
       id    int		  GENERATED ALWAYS AS IDENTITY,
       campaignid int		  NOT NULL,
       cmdrid	 int		  NOT NULL,
       FOREIGN KEY (campaignid) REFERENCES density.campaigns (id),
       FOREIGN KEY (cmdrid) REFERENCES density.cmdrs(id),
       PRIMARY KEY (id)
);
GRANT SELECT, INSERT ON density.surveys TO edservice;
GRANT SELECT ON density.surveys TO edviewer;

CREATE TABLE density.surveypoints (
       id    int		  GENERATED ALWAYS AS IDENTITY,
       surveyid int	  NOT NULL,
       sysname	     varchar(64)  NOT NULL,
       zsample	     int	  NOT NULL,
       x	     real	  NOT NULL,
       y	     real	  NOT NULL,
       z	     real	  NOT NULL,
       syscount	     int	  NOT NULL,
       maxdistance   real	  NOT NULL,
       FOREIGN KEY (surveyid) REFERENCES density.surveys(id),
       PRIMARY KEY (id),
       UNIQUE (surveyid, zsample),
       UNIQUE (surveyid, sysname),
       CHECK (syscount >= 0 AND syscount <= 50),
       CHECK (maxdistance > 0 AND maxdistance <= 20)
);
GRANT SELECT, INSERT ON density.surveypoints TO edservice;
GRANT SELECT ON density.surveypoints TO edviewer;
