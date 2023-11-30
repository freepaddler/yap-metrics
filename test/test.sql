
-- INSERT INTO metrics (name,type,f_value,updated_ts) VALUES ('g1','gauge',-117.3,now())
-- ON CONFLICT (name,type)
--     DO UPDATE SET f_value = excluded.f_value
-- RETURNING f_value;
--
--
-- INSERT INTO metrics (name,type,i_value,updated_ts) VALUES ('c1','counter',10,now())
-- ON CONFLICT (name,type)
--     DO UPDATE SET i_value = metrics.i_value + excluded.i_value
-- returning i_value;

drop table metrics;



