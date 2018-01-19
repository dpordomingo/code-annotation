const { compose } = require('react-app-rewired');
const rewireLess = require('react-app-rewire-less');

const rewireSass = function(config, env) {
  const oneOfRules = config.module.rules.find(r => r.oneOf)
  if (oneOfRules) {
    // Exclude sass files from the file-loader
    const fileLoaderRule = oneOfRules.oneOf.find(r => r.loader && r.loader.indexOf("file-loader") > -1);
    fileLoaderRule && fileLoaderRule.exclude.push(/\.scss$/);

    // Add a new rule to process sass
    oneOfRules.oneOf.push({
      test: /\.scss$/,
      loader: ["style-loader", "css-loader", "sass-loader"]
    });
  }

  return config;
}

module.exports = compose(rewireLess, rewireSass);
