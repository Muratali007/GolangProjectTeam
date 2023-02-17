ALTER TABLE footballers ADD CONSTRAINT footballers_year_check CHECK (year BETWEEN 1600 AND date_part('year', now()));
ALTER TABLE footballers ADD CONSTRAINT footballers_length_check CHECK (array_length(positions, 1) BETWEEN 1 AND 6);
