DROP INDEX IF EXISTS idx_movie_genres_genre_id;
DROP INDEX IF EXISTS idx_movie_genres_movie_id;
DROP INDEX IF EXISTS idx_reviews_rating;
DROP INDEX IF EXISTS idx_reviews_user_id;
DROP INDEX IF EXISTS idx_reviews_movie_id;
DROP INDEX IF EXISTS idx_movies_average_rating;
DROP INDEX IF EXISTS idx_movies_release_year;
DROP INDEX IF EXISTS idx_movies_title;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;

DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS movie_genres;
DROP TABLE IF EXISTS movies;
DROP TABLE IF EXISTS genres;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "uuid-ossp";

