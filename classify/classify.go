package classify

import (
	"encoding/gob"
	"os"
	"strconv"

	"github.com/meiji163/gh-spam/spam"
	"github.com/sjwhitworth/golearn/base"
)

var InstanceCols = []string{
	"association",
	"contributions",
	"repos",
	"age",
	"followers",
	"following",
	"body_len",
	"title_len",
	"sim_score",
	"is_spam",
}

// convert issue features to a golearn Instances object
func FeaturesToInstances(feats []spam.Features) *base.DenseInstances {
	attrs := make([]base.Attribute, len(InstanceCols))
	for i, col := range InstanceCols {
		attrs[i] = base.NewFloatAttribute(col)
		attrs[i].(*base.FloatAttribute).Precision = 0 // all ints
	}

	instances := base.NewDenseInstances()
	specs := make([]base.AttributeSpec, len(attrs))
	for i, attr := range attrs {
		specs[i] = instances.AddAttribute(attr)
	}

	instances.AddClassAttribute(attrs[len(attrs)-1])
	instances.Extend(len(feats))

	for row := 0; row < len(feats); row++ {
		feat := feats[row]
		vals := []int{
			feat.Association,
			feat.Contributions,
			feat.AuthorRepos,
			feat.AccountAge,
			feat.Followers,
			feat.Following,
			feat.BodyLen,
			feat.TitleLen,
			feat.TemplateScore,
			feat.IsSpam}

		for i := 0; i < len(specs); i++ {
			instances.Set(
				specs[i],
				row,
				specs[i].GetAttribute().GetSysValFromString(strconv.Itoa(vals[i])),
			)
		}
	}
	return instances
}

func WriteGob(filePath string, object interface{}) error {
	file, err := os.Create(filePath)
	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(object)
	}
	file.Close()
	return err
}

func ReadGob(filePath string, object interface{}) error {
	file, err := os.Open(filePath)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(object)
	}
	file.Close()
	return err
}
