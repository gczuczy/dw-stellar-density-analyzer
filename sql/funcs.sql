
CREATE OR REPLACE FUNCTION density.addsheetmeasurement(cmdr text, project text) RETURNS int AS $$
DECLARE
	cmdrid int;
	projectid int;
	mid int;
BEGIN
   SELECT INTO cmdrid id FROM density.cmdrs WHERE name = cmdr;
   IF NOT FOUND THEN
      INSERT INTO density.cmdrs (name) VALUES (cmdr) RETURNING id INTO cmdrid;
   END IF;

   SELECT INTO projectid id FROM density.projects WHERE name = project;
   IF NOT FOUND THEN
      INSERT INTO density.projects (name) VALUES (project)
      RETURNING id INTO projectid;
   END IF;

   INSERT INTO density.measurements (cmdrid, projectid) VALUES (cmdrid, projectid)
   RETURNING id INTO mid;

   RETURN mid;
END;
$$ LANGUAGE plpgsql VOLATILE STRICT PARALLEL UNSAFE SECURITY INVOKER;

GRANT EXECUTE ON FUNCTION density.addsheetmeasurement(cmdr text, project text) TO edservice;
